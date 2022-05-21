package flow

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/direktiv/direktiv/pkg/functions"
	grpcfunc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
)

func (srv *server) dependencyGraphPoller() {

	for {
		srv.updateDependencyGraph()
		// Dependency Polling Rate
		time.Sleep(time.Minute)
	}

}

type workflowDependencyGraph struct {
	Subflows        map[string]bool `json:"subflows"`
	Parents         map[string]bool `json:"parents"`
	NSFunctions     map[string]bool `json:"namespace_functions"`
	GlobalFunctions map[string]bool `json:"global_functions"`
	NSVars          map[string]bool `json:"namespace_variables"`
	Secrets         map[string]bool `json:"secrets"`
}

type secretsDependencyGraph struct {
	Workflows map[string]bool `json:"workflows"`
}

type nsvarDependencyGraph struct {
	Workflows map[string]bool `json:"workflows"`
}

type nsFunctionDependencyGraph struct {
	Workflows map[string]bool `json:"workflows"`
}

type namespaceDependencyGraph struct {
	ID          string                                `json:"-"`
	Workflows   map[string]*workflowDependencyGraph   `json:"workflows"`
	Secrets     map[string]*secretsDependencyGraph    `json:"secrets"`
	Vars        map[string]*nsvarDependencyGraph      `json:"variables"`
	NSFunctions map[string]*nsFunctionDependencyGraph `json:"namespace_functions"`
}

type wfDependency struct {
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
}

type globalFunctionNamespaceDependenciesGraph struct {
	Workflows map[string]bool `json:"workflows"`
}

type globalFunctionDependencyGraph struct {
	Namespaces map[string]*globalFunctionNamespaceDependenciesGraph `json:"namespaces"`
}

type dependencyGraph struct {
	Namespaces      map[string]*namespaceDependencyGraph      `json:"namespaces"`
	GlobalFunctions map[string]*globalFunctionDependencyGraph `json:"global_functions"`
}

func (srv *server) scrapeWorkflows() (*dependencyGraph, error) {

	dg := new(dependencyGraph)
	dg.Namespaces = make(map[string]*namespaceDependencyGraph)

	ctx := context.Background()

	namespaces, err := srv.db.Namespace.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces {

		ndg := new(namespaceDependencyGraph)
		ndg.ID = namespace.ID.String()
		ndg.Workflows = make(map[string]*workflowDependencyGraph)
		ndg.Secrets = make(map[string]*secretsDependencyGraph)
		ndg.Vars = make(map[string]*nsvarDependencyGraph)
		ndg.NSFunctions = make(map[string]*nsFunctionDependencyGraph)

		dg.Namespaces[namespace.Name] = ndg

		workflows, err := namespace.QueryWorkflows().All(ctx)
		if err != nil {
			srv.sugar.Error("Dependency graph update error: %v", err)
			continue
		}

		for _, workflow := range workflows {

			wfd, err := srv.reverseTraverseToWorkflow(ctx, workflow.ID.String())
			if err != nil {
				srv.sugar.Error("Dependency graph update error: %v", err)
				continue
			}

			wdg := new(workflowDependencyGraph)
			wdg.Subflows = make(map[string]bool)
			wdg.Parents = make(map[string]bool)
			wdg.GlobalFunctions = make(map[string]bool)
			wdg.NSFunctions = make(map[string]bool)
			wdg.NSVars = make(map[string]bool)
			wdg.Secrets = make(map[string]bool)

			ndg.Workflows[GetInodePath(wfd.path)] = wdg

			revisions, err := workflow.QueryRevisions().All(ctx)
			if err != nil {
				srv.sugar.Error("Dependency graph update error: %v", err)
				continue
			}

			for _, revision := range revisions {

				wf := new(model.Workflow)

				err = wf.Load(revision.Source)
				if err != nil {
					continue
				}

				for _, fn := range wf.Functions {

					switch fn.(type) {
					case *model.SubflowFunctionDefinition:
						sfd := fn.(*model.SubflowFunctionDefinition)
						wdg.Subflows[GetInodePath(sfd.Workflow)] = false
					case *model.ReusableFunctionDefinition:
						// rfd := fn.(*model.ReusableFunctionDefinition)
						// TODO
					case *model.NamespacedFunctionDefinition:
						nfd := fn.(*model.NamespacedFunctionDefinition)
						wdg.NSFunctions[nfd.KnativeService] = false
						// TODO
					case *model.GlobalFunctionDefinition:
						gfd := fn.(*model.GlobalFunctionDefinition)
						wdg.GlobalFunctions[gfd.KnativeService] = false
					default:
						srv.sugar.Error("Dependency graph update error: wrong type %v", reflect.TypeOf(fn))
					}

				}

				for _, state := range wf.States {

					switch state.(type) {
					case *model.SetterState:
					case *model.GetterState:
					case *model.ActionState:
						as := state.(*model.ActionState)
						for _, secret := range as.Action.Secrets {
							wdg.Secrets[secret] = false
						}

						for _, file := range as.Action.Files {
							switch file.Scope {
							case "namespace":
								wdg.NSVars[file.Key] = false
							case "workflow":
							case "":
								fallthrough
							case "instance":
							case "inline":
							default:
								srv.sugar.Error("Dependency graph update error: wrong scope %v", file.Scope)
							}
						}

					case *model.ParallelState:
						ps := state.(*model.ParallelState)
						for _, action := range ps.Actions {
							for _, secret := range action.Secrets {
								wdg.Secrets[secret] = false
							}

							for _, file := range action.Files {
								switch file.Scope {
								case "namespace":
									wdg.NSVars[file.Key] = false
								case "workflow":
								case "":
									fallthrough
								case "instance":
								case "inline":
								default:
									srv.sugar.Error("Dependency graph update error: wrong scope %v", file.Scope)
								}
							}
						}

					case *model.ForEachState:
						fs := state.(*model.ForEachState)
						for _, secret := range fs.Action.Secrets {
							wdg.Secrets[secret] = false
						}

						for _, file := range fs.Action.Files {
							switch file.Scope {
							case "namespace":
								wdg.NSVars[file.Key] = false
							case "workflow":
							case "":
								fallthrough
							case "instance":
							case "inline":
							default:
								srv.sugar.Error("Dependency graph update error: wrong scope %v", file.Scope)
							}
						}

					default:
					}

				}

			}

		}

	}

	return dg, nil

}

func (srv *server) resolveDependencyGraphLinks(dg *dependencyGraph) error {

	ctx := context.Background()

	lfresp, err := srv.actions.client.ListFunctions(ctx, &grpcfunc.ListFunctionsRequest{
		Annotations: map[string]string{
			functions.ServiceHeaderScope: functions.PrefixGlobal,
		},
	})
	dg.GlobalFunctions = make(map[string]*globalFunctionDependencyGraph)
	if err != nil {
		srv.sugar.Error("Dependency graph update error: %v", err)
	} else {
		for _, fn := range lfresp.Functions {
			s := fn.GetInfo().GetName()
			gfdg := new(globalFunctionDependencyGraph)
			gfdg.Namespaces = make(map[string]*globalFunctionNamespaceDependenciesGraph)
			dg.GlobalFunctions[s] = gfdg
		}
	}

	// TODO: namespace services

	for ns, ndg := range dg.Namespaces {

		// global services

		for wf, wdg := range ndg.Workflows {

			var globals = make([]string, 0)

			for global := range wdg.GlobalFunctions {
				globals = append(globals, global)
			}

			for _, global := range globals {

				gfdg, exists := dg.GlobalFunctions[global]
				if exists {
					wdg.GlobalFunctions[global] = true
					gfdgns, exists := gfdg.Namespaces[ns]
					if !exists {
						gfdgns = new(globalFunctionNamespaceDependenciesGraph)
						gfdgns.Workflows = make(map[string]bool)
						gfdg.Namespaces[ns] = gfdgns
					}

					gfdgns.Workflows[wf] = true

				}

			}

		}

		// namespace services

		nsresp, err := srv.actions.client.ListFunctions(ctx, &grpcfunc.ListFunctionsRequest{
			Annotations: map[string]string{
				functions.ServiceHeaderNamespaceID: ndg.ID,
				functions.ServiceHeaderScope:       functions.PrefixNamespace,
			},
		})
		if err != nil {
			srv.sugar.Error("Dependency graph update error: %v", err)
		} else {
			for _, fn := range nsresp.Functions {
				s := fn.GetInfo().GetName()
				nsfdg := new(nsFunctionDependencyGraph)
				nsfdg.Workflows = make(map[string]bool)
				ndg.NSFunctions[s] = nsfdg
			}
		}

		for wf, wdg := range ndg.Workflows {

			var nsfns = make([]string, 0)

			for nsfn := range wdg.NSFunctions {
				nsfns = append(nsfns, nsfn)
			}

			for _, nsfn := range nsfns {

				nsfndg, exists := ndg.NSFunctions[nsfn]
				if exists {
					wdg.NSFunctions[nsfn] = true
					nsfndg.Workflows[wf] = true
				}

			}

		}

		// ns vars
		x, err := srv.getNamespace(ctx, srv.db.Namespace, ns)
		if err != nil {
			srv.sugar.Error("Dependency graph update error: %v", err)
			continue
		}

		vrefs, err := x.QueryVars().All(ctx)
		if err != nil {
			srv.sugar.Error("Dependency graph update error: %v", err)
			continue
		}

		for _, vref := range vrefs {
			ndg.Vars[vref.Name] = &nsvarDependencyGraph{
				Workflows: make(map[string]bool),
			}
		}

		for wf, wdg := range ndg.Workflows {

			var vars = make([]string, 0)

			for v := range wdg.NSVars {
				vars = append(vars, v)
			}

			for _, v := range vars {

				vdg, exists := ndg.Vars[v]
				if exists {
					wdg.NSVars[v] = true
					vdg.Workflows[wf] = true
				}

			}

		}

		// secrets
		resp, err := srv.secrets.client.GetSecrets(ctx, &secretsgrpc.GetSecretsRequest{
			Namespace: &ndg.ID,
		})
		if err != nil {
			srv.sugar.Error("Dependency graph update error: %v", err)
			continue
		}

		for _, secret := range resp.Secrets {
			s := secret.GetName()
			ndg.Secrets[s] = &secretsDependencyGraph{
				Workflows: make(map[string]bool),
			}
		}

		for wf, wdg := range ndg.Workflows {

			var secrets = make([]string, 0)

			for secret := range wdg.Secrets {
				secrets = append(secrets, secret)
			}

			for _, secret := range secrets {

				sdg, exists := ndg.Secrets[secret]
				if exists {
					wdg.Secrets[secret] = true
					sdg.Workflows[wf] = true
				}

			}

		}

		// subflows
		for wf, wdg := range ndg.Workflows {

			var sfs = make([]string, 0)

			for sf := range wdg.Subflows {
				sfs = append(sfs, sf)
			}

			for _, sf := range sfs {
				child, exists := ndg.Workflows[sf]
				if exists {
					wdg.Subflows[sf] = true
					child.Parents[wf] = true
				}
			}

		}

	}

	return nil

}

func (srv *server) updateDependencyGraph() {

	srv.sugar.Debug("Updating dependency graphs...")

	dg, err := srv.scrapeWorkflows()
	if err != nil {
		srv.sugar.Error("Dependency graph update error: %v", err)
		return
	}

	err = srv.resolveDependencyGraphLinks(dg)
	if err != nil {
		srv.sugar.Error("Dependency graph update error: %v", err)
		return
	}

	srv.dependencies = dg

	srv.sugar.Debug("Dependency graphs updated.")

	// data := marshal(dg)
	// fmt.Println(data)

}

func (srv *server) getCompleteDependencyGraph() (*dependencyGraph, error) {

	if srv.dependencies == nil {
		return nil, errors.New("no cached dependency graph")
	}

	return srv.dependencies, nil

}

func (srv *server) getNamespacedDependencyGraph(ns string) (*namespaceDependencyGraph, error) {

	if srv.dependencies == nil {
		return nil, errors.New("no cached dependency graph")
	}

	full := srv.dependencies

	nsdg, exists := full.Namespaces[ns]
	if !exists {
		return nil, errors.New("no cached dependency graph for namespace")
	}

	return nsdg, nil

}
