package secrets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	annotationNamespace = "direktiv.io/namespace"
	annotationName      = "direktiv.io/name"
	secretKey           = "secret"
)

var (
	ErrNotFound = errors.New("ErrNotFound")
	nameRegex   = regexp.MustCompile(`^[a-z\-]{1,24}$`)
)

type Manager struct {
	cache     cache.Cache[core.Secret]
	clientSet kubernetes.Interface
	namespace string
}

func NewManager(c *core.Config, cache cache.Cache[core.Secret]) (core.SecretsManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("failed to get in-cluster config", slog.Any("error", err))
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create clientset", slog.Any("error", err))
		return nil, err
	}

	return &Manager{
		cache:     cache,
		namespace: c.DirektivNamespace,
		clientSet: clientSet,
	}, nil
}

func (sm *Manager) Get(ctx context.Context, namespace, name string) (*core.Secret, error) {
	s, err := sm.cache.Get(toKubernetesName(namespace, name), func(a ...any) (core.Secret, error) {
		s, err := sm.getKubernetesSecret(ctx, namespace, name)
		if err != nil {
			return core.Secret{}, err
		}

		return core.Secret{
			Name:      s.Name,
			CreatedAt: s.CreationTimestamp.Time,
			Data:      s.Data[secretKey],
		}, nil
	})

	return &s, err
}

func (sm *Manager) Create(ctx context.Context, namespace string, secret *core.Secret) (*core.Secret, error) {
	slog.Info("creating secret", slog.String("secret", secret.Name))

	kname := toKubernetesName(namespace, secret.Name)

	if !nameRegex.MatchString(secret.Name) {
		slog.Error("creating secret failed because of an invalid name")
		return nil, fmt.Errorf("invalid secret name")
	}

	// Validate value is not empty
	if string(secret.Data) == "" {
		slog.Error("creating secret failed because it is empty")
		return nil, fmt.Errorf("secret value cannot be empty")
	}

	secretKubernetes := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kname,
			Namespace: sm.namespace,
			Labels: map[string]string{
				annotationNamespace: namespace,
				annotationName:      secret.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{secretKey: secret.Data},
	}

	s, err := sm.clientSet.CoreV1().Secrets(sm.namespace).Create(
		ctx,
		secretKubernetes,
		metav1.CreateOptions{},
	)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			slog.Error("creating secret failed because it already exists")
			return nil, datastore.ErrDuplication
		}

		return nil, err
	}

	return &core.Secret{
		Name:      secret.Name,
		CreatedAt: s.CreationTimestamp.Time,
		Data:      secret.Data,
	}, nil
}

func (sm *Manager) GetAll(ctx context.Context, namespace string) ([]*core.Secret, error) {
	s := make([]*core.Secret, 0)
	secretItems, err := sm.clientSet.CoreV1().Secrets(sm.namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace),
		},
	)
	if err != nil {
		return nil, err
	}

	for i := range secretItems.Items {
		s = append(s, &core.Secret{
			Name:      secretItems.Items[i].Labels[annotationName],
			CreatedAt: secretItems.Items[i].CreationTimestamp.Time,
			Data:      secretItems.Items[i].Data[secretKey],
		})
	}

	return s, nil
}

func (sm *Manager) Update(ctx context.Context, namespace string, secret *core.Secret) (*core.Secret, error) {
	slog.Info("updating secret", slog.String("secret", secret.Name))
	s, err := sm.getKubernetesSecret(ctx, namespace, secret.Name)
	if err != nil {
		slog.Error("updating secret failed", slog.Any("secret", err.Error()))
		return nil, err
	}

	s.Data = map[string][]byte{secretKey: secret.Data}

	_, err = sm.clientSet.CoreV1().Secrets(sm.namespace).Update(
		ctx,
		s,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return nil, err
	}

	sm.cache.Notify(ctx, cache.CacheNotify{
		Key:    toKubernetesName(namespace, secret.Name),
		Action: cache.CacheUpdate,
	})

	return secret, nil
}

func (sm *Manager) Delete(ctx context.Context, namespace, name string) error {
	slog.Info("deleting secret", slog.String("secret", name))
	kname := toKubernetesName(namespace, name)

	err := sm.clientSet.CoreV1().Secrets(sm.namespace).Delete(ctx, kname, metav1.DeleteOptions{})
	if apierrors.ReasonForError(err) == metav1.StatusReasonNotFound {
		slog.Error("deleting secret failed", slog.Any("secret", err.Error()))
		return datastore.ErrNotFound
	}

	// only update if dleete was successful
	if err == nil {
		sm.cache.Notify(ctx, cache.CacheNotify{
			Key:    toKubernetesName(namespace, name),
			Action: cache.CacheDelete,
		})
	}

	return err
}

func (sm *Manager) DeleteForNamespace(ctx context.Context, namespace string) error {
	slog.Info("deleting secrets in a namespace", slog.String("namespace", namespace))

	secrets, err := sm.GetAll(ctx, namespace)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		err := sm.Delete(ctx, namespace, secret.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sm *Manager) getKubernetesSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	kname := toKubernetesName(namespace, name)

	secret, err := sm.clientSet.CoreV1().Secrets(sm.namespace).Get(ctx, kname, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

func toKubernetesName(namespace, name string) string {
	namespace = strings.ReplaceAll(namespace, "_", "-")
	return fmt.Sprintf("nssecret-%s-%s", namespace, name)
}
