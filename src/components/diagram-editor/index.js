import { useGlobalServices, useNamespaceServices, useNodes } from 'direktiv-react-hooks';
import { useCallback, useEffect, useState } from 'react';
import { VscGear, VscListUnordered, VscSymbolEvent, VscInfo, VscFileCode } from 'react-icons/vsc';
import Alert from '../../components/alert';
import FlexBox from '../../components/flexbox';
import { Config } from '../../util';
import Drawflow from 'drawflow';
import { Resizable } from 're-resizable';
import { DefaultSchemaUI, GenerateFunctionSchemaWithEnum, getSchemaCallbackMap, getSchemaDefault, SchemaUIMap } from "../../components/diagram-editor/jsonSchema"
import Form from '@rjsf/core';
import { CreateNode, DefaultValidateSubmitCallbackMap, nodeGetInputConnections, onSubmitCallbackMap, onValidateSubmitCallbackMap, setConnections, sortNodes, unescapeJSStrings } from '../../components/diagram-editor/util';
import { AutoSizer, CellMeasurer, CellMeasurerCache, List } from 'react-virtualized';
import Fuse from 'fuse.js';
import { ActionsNodes, NodeStateAction } from "../../components/diagram-editor/nodes";
import PrettyYAML from "json-to-pretty-yaml"

// Import Styles
import './styles/form.css';
import './styles/node.css';
import './styles/style.css';
import 'drawflow/dist/drawflow.min.css'

import { importFromYAML } from '../../components/diagram-editor/import';
import Modal, { ButtonDefinition, ModalHeadless } from '../modal';

import Ajv from "ajv"
import { CustomWidgets } from './widgets';

const actionsNodesFuse = new Fuse(ActionsNodes, {
    keys: ['name']
})


function Actions(props) {
    const cache = new CellMeasurerCache({
        fixedWidth: false,
        fixedHeight: true
    })

    function rowRenderer({ index, parent, key, style }) {
        return (
            <CellMeasurer
                key={`action-${key}`}
                cache={cache}
                parent={parent}
                columnIndex={0}
                rowIndex={index}
            >
                <div style={{ ...style, minHeight: "90px", height: "90px", cursor: "move", userSelect: "none", display: "flex" }}>
                    <div className={`action ${ActionsNodes[index].family} action-${ActionsNodes[index].type}`} draggable={true} node-index={index} onDragStart={(ev) => {
                        ev.stopPropagation();
                        ev.dataTransfer.setData("nodeIndex", ev.target.getAttribute("node-index"));
                    }}>
                        <div style={{ marginLeft: "5px", marginRight: "2px" }}>
                            <div style={{ display: "flex", borderBottom: "1px solid #e5e5e5", justifyContent: "space-between" }}>
                                <span style={{ whiteSpace: "pre-wrap", cursor: "move", fontSize: "13px", overflow: "hidden", textOverflow: "ellipsis" }}>
                                    {ActionsNodes[index].name}
                                </span>
                                {
                                    ActionsNodes[index].info.link ?
                                    <a style={{ whiteSpace: "nowrap", cursor: "pointer", fontSize: "11px", paddingRight: "3px", display: "flex", alignItems: "center", justifyContent: "center" }} href={`${ActionsNodes[index].info.link}`} target="_blank" rel="noreferrer">
                                        <VscInfo />
                                    </a>
                                    :
                                    <></>
                                }
                            </div>
                            <div style={{ fontSize: "10px", lineHeight: "10px", paddingTop: "2px" }}>
                                <p style={{ whiteSpace: "pre-wrap", cursor: "move", margin: "0px" }}>
                                    {ActionsNodes[index].info.description}
                                </p>
                            </div>
                        </div>

                    </div>
                </div>
            </CellMeasurer>
        );
    }

    return (
        <AutoSizer>
            {({ height, width }) => (
                <div style={{ height: "100%", minHeight: "100%" }}>
                    <List
                        width={width}
                        height={height}
                        rowRenderer={rowRenderer}
                        deferredMeasurementCache={cache}
                        rowCount={ActionsNodes.length}
                        rowHeight={90}
                        scrollToAlignment={"start"}
                    />
                </div>
            )}
        </AutoSizer>
    )
}

function FunctionsList(props) {
    const { functionList, setFunctionList, namespace, functionDrawerWidth } = props

    const [newFunctionFormRef, setNewFunctionFormRef] = useState(null)
    const [formData, setFormData] = useState({})

    const namespaceServiceHook = useNamespaceServices(Config.url, false, namespace, localStorage.getItem("apikey"))
    const globalServiceHook = useGlobalServices(Config.url, false, localStorage.getItem("apikey"))
    const namespaceNodesHook = useNodes(Config.url, false, namespace, "/", localStorage.getItem("apikey"), "first=20")


    if (namespaceServiceHook.data === null || globalServiceHook.data === null || namespaceNodesHook.data === null) {
        return <></>
    }

    const ajv = new Ajv()
    const functionSchemas = GenerateFunctionSchemaWithEnum(namespaceServiceHook.data.map(a => a.serviceName), (globalServiceHook.data.map(a => a.serviceName)), namespaceNodesHook.data)
    const validate = ajv.compile(functionSchemas.schema)

    // Set uischema to global or namespace depending on the type
    // It does not matter if global uiSchema is used for reusable or subflow since the dont have the service field
    let uiSchema = functionSchemas.uiSchema['knative-global']
    if (functionSchemas.schema && formData && formData.type === "knative-namespace") {
        uiSchema = functionSchemas.uiSchema['knative-namespace']
    }

    const cache = new CellMeasurerCache({
        fixedWidth: false,
        fixedHeight: true
    })

    function rowRenderer({ index, parent, key, style }) {
        return (
            <CellMeasurer
                key={key}
                cache={cache}
                parent={parent}
                columnIndex={0}
                rowIndex={index}
            >
                <div style={{ ...style, minHeight: "6px", height: "84px", cursor: "move", userSelect: "none", display: "flex" }}>
                    <div className={`function`} draggable={true} function-index={index} onDragStart={(ev) => {
                        ev.dataTransfer.setData("functionIndex", ev.target.getAttribute("function-index"));
                    }}>
                        <div class="node-labels" style={{ display: "flex", gap: "4px", flexDirection: "column", marginLeft: "5px" }}>
                            <div>
                                ID: <span class="label-id" style={{ maxWidth: functionDrawerWidth - 50 }}>{functionList[index].id}</span>
                            </div>
                            <div>
                                Type: <span class="label-type">{functionList[index].type}</span>
                            </div>
                            <div>
                                {functionList[index].service ? `Service:` : ""}
                                {functionList[index].image ? `Image:` : ""}
                                {functionList[index].workflow ? `Workflow:` : ""}
                                <span style={{ maxWidth: functionDrawerWidth - 80 }} class="label-type">
                                    {functionList[index].service ? `${functionList[index].service}` : ""}
                                    {functionList[index].image ? `${functionList[index].image}` : ""}
                                    {functionList[index].workflow ? `${functionList[index].workflow}` : ""}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </CellMeasurer>
        );
    }

    return (
        <>
            <Modal
                style={{ justifyContent: "center" }}
                className="run-workflow-modal"
                modalStyle={{ color: "black", width: "600px" }}
                title={`Create Function`}
                onClose={() => {
                    setFormData({})
                }}
                actionButtons={[
                    ButtonDefinition("Create Function", async () => {
                        newFunctionFormRef.click()

                        // Check if form data is valid
                        if (!validate(formData)) {
                            if (formData.type === "knative-namespace" && namespaceServiceHook.data.length === 0) {
                                throw Error("Invalid Function: No Namespace Services Exist")
                            } else if (formData.type === "knative-global" && globalServiceHook.data.length === 0) {
                                throw Error("Invalid Function: No Global Services Exist")
                            }

                            throw Error("Invalid Function")
                        }

                        // Throw error if id already exists
                        const result = functionList.filter(functionItem => functionItem.id === formData.id)
                        if (result.length > 0) {
                            throw Error(`Function '${formData.id}' already exists`)
                        }

                        // Update list
                        setFunctionList((oldfList) => {
                            oldfList.push(formData)
                            return [...oldfList]
                        })
                    }, "small blue", () => { }, true, false, true),
                    ButtonDefinition("Cancel", async () => {
                    }, "small light", () => { }, true, false)
                ]}
                button={(
                    <div className={`btn function-btn`}>
                        New function
                    </div>

                )}
            >
                <FlexBox className="col" style={{ height: "45vh", minWidth: "250px", minHeight: "200px", justifyContent: "space-between" }}>
                    <div style={{ overflow: "auto" }}>
                        <Form
                            id={"builder-form"}
                            onSubmit={(form) => {
                            }}
                            schema={functionSchemas.schema}
                            uiSchema={uiSchema}
                            formData={formData}
                            widgets={CustomWidgets}
                            onChange={(e) => {
                                setFormData(e.formData)
                            }}
                        >
                            <button ref={setNewFunctionFormRef} style={{ display: "none" }} />
                        </Form>
                    </div>
                </FlexBox>
            </Modal>
            {functionList.length > 0 ? (
                <AutoSizer>
                    {({ height, width }) => (
                        <List
                            width={width}
                            height={height}
                            rowRenderer={rowRenderer}
                            deferredMeasurementCache={cache}
                            rowCount={functionList.length}
                            rowHeight={84}
                            scrollToAlignment={"start"}
                        />
                    )}
                </AutoSizer>)
                :
                (<>
                    <div style={{ display: "flex", justifyContent: "center" }}>
                        No functions
                    </div>
                </>)}
        </>
    )
}

const MaxDrawerSize = 180

export default function DiagramEditor(props) {
    const { workflow, namespace, updateWorkflow, setBlock } = props

    const [diagramEditor, setDiagramEditor] = useState(null);
    const [load, setLoad] = useState(true);

    const [selectedNode, setSelectedNode] = useState(null);
    const [formRef, setFormRef] = useState(null);
    const [error, setError] = useState(null)

    const [actionDrawerWidth, setActionDrawerWidth] = useState(0)
    const [actionDrawerWidthOld, setActionDrawerWidthOld] = useState(MaxDrawerSize)
    const [actionDrawerMinWidth, setActionDrawerMinWidth] = useState(0)

    const [unhandledData, setUnhandledData] = useState({})

    const [functionDrawerWidth, setFunctionDrawerWidth] = useState(0)
    const [functionDrawerMinWidth, setFunctionDrawerMinWidth] = useState(0)
    const [functionList, setFunctionList] = useState([])

    const [nodeDetailsVisible, setNodeDetailsVisible] = useState(false)
    const [nodeIDModalVisible, setNodeIDModalVisible] = useState(false)
    const [newNodeID, setNewNodeID] = useState("")
    const [selectedNodeFormData, setSelectedNodeFormData] = useState({})
    const [oldSelectedNodeFormData, setOldSelectedNodeFormData] = useState({})
    const [selectedNodeSchema, setSelectedNodeSchema] = useState({})
    const [selectedNodeSchemaUI, setSelectedNodeSchemaUI] = useState({})

    // Track whether all nodes have been init'd
    // If all nodes are init (valid form submited) then diagram can compile
    const [nodeInitTracker, setNodeInitTracker] = useState({})
    // const [canCompile, setCanCompile] = useState(true)


    // Context menu to add nodes
    const [contextMenuAnchorPoint, setContextMenuAnchorPoint] = useState({ x: 0, y: 0 });

    const [showContextMenu, setShowContextMenu] = useState(false);
    const [contextMenuResults, setContextMenuResults] = useState(ActionsNodes)

    // Context menu to edit nodes
    const [showNodeContextMenu, setShowNodeContextMenu] = useState(false);

    useEffect(() => {
        if (selectedNode) {
            const getSchema = getSchemaCallbackMap[selectedNode.data.schemaKey]
            if (getSchema) {
                setSelectedNodeSchema(getSchema(selectedNode.data.schemaKey, functionList))
            } else {
                setSelectedNodeSchema(getSchemaDefault(selectedNode.data.schemaKey))
            }

            //  Apply custom schema UI
            //  https://react-jsonschema-form.readthedocs.io/en/latest/usage/widgets/
            const schemaUI = SchemaUIMap[selectedNode.data.schemaKey]
            if (schemaUI) {
                setSelectedNodeSchemaUI(schemaUI)
            } else {
                // Clear UI schema  
                setSelectedNodeSchemaUI(DefaultSchemaUI)
            }
        }
    }, [selectedNode, functionList])

    useEffect(() => {
        var id = document.getElementById("drawflow");
        // if (!diagramEditor) {
        let editor = new Drawflow(id)
        editor.start()
        editor.force_first_input = true
        editor.on('nodeSelected', function (id) {
            const node = editor.getNodeFromId(id)
            setSelectedNode(node)
            setSelectedNodeFormData(node.data.formData)
            setOldSelectedNodeFormData(node.data.formData)
        })

        editor.on("connectionCreated", function (ev) {
            // Handle Special cases where we need to remove or adjust connections
            // INFO: Connections are created from output to input
            // E.g. {[input_1]NodeA[output_1]}--->{[input_1]NodeB[output_1]}
            const outNode = editor.getNodeFromId(ev.output_id)
            const inNode = editor.getNodeFromId(ev.input_id)
            let errorOutput = 'output_2'
            const outputIsPrimitive = outNode.class.includes("primitive")
            let isInvalidConnection = false

            // XOR has no default transition so output_1 will be used for errors
            // TODO: It might be worth generating error output class from "output" + node.info.output
            if (outNode.name === "StateEventXor") {
                errorOutput = "output_1"
            }

            if (outputIsPrimitive) {
                if (ev.output_class === errorOutput && inNode.name !== "CatchError") {
                    // Remove connection if pirimtive node Erorr output is going to non-errorblock
                    isInvalidConnection = true
                } else if (ev.output_class !== errorOutput && inNode.name === "CatchError") {
                    // Remove connection if pirimtive node transition output is going to errorblock
                    isInvalidConnection = true
                }
            } else if (inNode.name === "CatchError" && !outputIsPrimitive) {
                // Remove connection in input is error block, but output is primitive
                isInvalidConnection = true
            }

            if (isInvalidConnection) {
                editor.removeSingleConnection(ev.output_id, ev.input_id, ev.output_class, ev.input_class)
            } else {
                // If output connection already existed before creation, delete first connection and keep new one.
                if (outNode.outputs[ev.output_class].connections.length > 1) {
                    const removeConnection = outNode.outputs[ev.output_class].connections[0];
                    editor.removeSingleConnection(ev.output_id, removeConnection.node, ev.output_class, removeConnection.output);
                }
            }

        })

        editor.on('nodeCreated', function (id) {
            let node = editor.getNodeFromId(id)
            // If node was created without id, geneate one
            if (!node.data.id) {
                if (node.data.family === "special") {
                    // Manually set id if its a special block
                    // Since this is not a state, we dont need unique ids for special
                    node.data.id = `${node.data.type}-block`
                } else {
                    node.data.id = `node-${id}-${node.data.type}`
                }
                editor.updateNodeDataFromId(id, node.data)
            }

            // Track node init state
            setNodeInitTracker((old) => {
                old[id] = { init: node.data.init, stateID: node.data.id }
                return {
                    ...old
                }
            })
        })

        editor.on('nodeUnselected', function (e) {
            setSelectedNode(null)
        })

        editor.on('nodeRemoved', function (e) {
            setSelectedNode(null)
        })

        setDiagramEditor(editor)
    }, [])

    // Import if diagram editor is mounted and workflow was passed in props
    useEffect(() => {
        if (diagramEditor && workflow) {
            if (load) {
                setUnhandledData(importFromYAML(diagramEditor, setFunctionList, workflow))
            }
            setLoad(false)
        }
    }, [diagramEditor, workflow, load])

    const resizeStyle = {
        display: "flex",
        alignItems: "center",
        justifyContent: "flex-start",
        flexDirection: "column",
        background: "#f0f0f0",
        zIndex: 10,
    };

    // Update Context Menu results
    useEffect(() => {
        if (!showContextMenu) {
            setContextMenuResults(ActionsNodes)
        }
    }, [showContextMenu]);


    // Hide Context Menu if user clicks somewhere else
    // Show node context menu when clicking on node `node-btn` div
    // FIXME: Might cause problems for users who try to click on context-menu searchbar
    const handleClick = useCallback((ev) => {
        if (showContextMenu) {
            setShowContextMenu(false)
        }

        if (ev && ev.target && ev.target.id) {
            switch (ev.target.id) {
                case "node-btn":
                    // Edit button was clicked on node
                    // We can assume that a node is selected
                    setContextMenuAnchorPoint({ x: ev.pageX, y: ev.pageY })
                    setShowNodeContextMenu(true)
                    break;
                default:
                    break;
            }
        } else if (showNodeContextMenu) {
            setShowNodeContextMenu(false)
        }


    }, [showContextMenu, showNodeContextMenu]);
    useEffect(() => {
        document.addEventListener("click", handleClick);
        return () => {
            document.removeEventListener("click", handleClick);
        };
    });

    return (
        <>
            {showContextMenu ? (
                <div
                    id='context-menu'
                    className="context-menu"
                    style={{
                        top: contextMenuAnchorPoint.y,
                        left: contextMenuAnchorPoint.x
                    }}
                >
                    <div style={{ textAlign: "center", padding: "4px 2px 4px 2px", fontWeight:"bold"}}>
                        Add Node
                    </div>
                    <input autoFocus type="search" id="fname" name="fname" onChange={(ev) => {
                        setContextMenuResults(actionsNodesFuse.search(ev.target.value))
                    }}
                        onKeyDown={(ev) => {
                            if (ev.key === 'Enter' && contextMenuResults.length > 0) {
                                const newNode = contextMenuResults[0].item ? contextMenuResults[0].item : contextMenuResults[0]
                                setBlock(true)
                                CreateNode(diagramEditor, newNode, contextMenuAnchorPoint.x, contextMenuAnchorPoint.y)

                                setShowContextMenu(false)
                                setShowNodeContextMenu(false)

                            }
                        }}
                    ></input>
                    <ul >
                        {
                            contextMenuResults.map((obj) => {
                                return (
                                    <li onClick={() => {
                                        const newNode = obj.item ? obj.item : obj
                                        setBlock(true)
                                        CreateNode(diagramEditor, newNode, contextMenuAnchorPoint.x, contextMenuAnchorPoint.y)
                                        setShowContextMenu(false)
                                        setShowNodeContextMenu(false)
                                    }}>
                                        {obj.name ? obj.name : obj.item.name}
                                    </li>
                                )
                            })
                        }
                    </ul>
                </div>
            ) : (
                <> </>
            )}
            {showNodeContextMenu ? (
                <div
                    id='context-menu'
                    className="context-menu"
                    style={{
                        top: contextMenuAnchorPoint.y,
                        left: contextMenuAnchorPoint.x
                    }}
                >
                    <div style={{ textAlign: "center", padding: "4px 2px 4px 2px", fontWeight:"bold" }}>
                        Node Options
                    </div>
                    <ul >
                        <li onClick={() => {
                            setNodeDetailsVisible(true)
                        }}>
                            Edit Values
                        </li>
                        {/* Only show delete option if selected node is not a start node */}
                        {selectedNode && selectedNode.data.family !== "special" ?
                            <li onClick={() => {
                                setNodeIDModalVisible(true)
                            }}>
                                Edit ID
                            </li>
                            : <></>
                        }
                        {/* Only show delete option if selected node is not a start node */}
                        {selectedNode && selectedNode.data.type !== "start" ?
                            <li onClick={() => {
                                diagramEditor.removeNodeId(`node-${selectedNode.id}`)
                                setNodeInitTracker((old) => {
                                    delete old[selectedNode.id]
                                    return {
                                        ...old
                                    }
                                })
                            }}>
                                Delete
                            </li>
                            : <></>
                        }

                    </ul>
                </div>
            ) : (
                <> </>
            )}
            <FlexBox id="builder-page" className="col" style={{ paddingRight: "8px" }}>
                {error ?
                    <Alert className="critical" style={{ flex: "0", margin: "3px" }}>{error} </Alert>
                    :
                    <></>
                }
                {/* <div style={{height:"600px", width: "600px"}}> */}
                <div className='toolbar'>
                    <div className='toolbar-btn' onClick={() => {
                        setError(null)

                        // Check if any nodes are have not been initialized
                        let nonInitNodes = []
                        for (const n of Object.keys(nodeInitTracker)) {
                            if (!nodeInitTracker[`${n}`].init) {
                                nonInitNodes.push(nodeInitTracker[`${n}`].stateID)
                            }
                        }

                        if (nonInitNodes.length > 0) {
                            setError(`Failed - Node not initialized: ${nonInitNodes.join(", ")}`)
                            return
                        }


                        // Export Nodes to an object so we can convert it to a workflow yaml
                        let rawExport = diagramEditor.export()
                        let rawData = rawExport.drawflow.Home.data
                        let wfData = { start: {}, functions: functionList, states: [] }

                        // Delete empty functions from workflow YAML
                        if (functionList.length === 0) {
                            delete wfData["functions"]
                        }

                        // Find Start Block
                        const startBlockIDs = diagramEditor.getNodesFromName("StartBlock")
                        rawData[startBlockIDs[0]].compiled = true
                        let startBlock = rawData[startBlockIDs[0]];
                        let startState

                        // Setup Workflow Start State
                        wfData.start.type = startBlock.data.formData.type
                        switch (startBlock.data.formData.type) {
                            case "default":
                                break;
                            case "scheduled":
                                wfData.start.cron = startBlock.data.formData.cron
                                break;
                            case "event":
                                wfData.start.event = startBlock.data.formData.event
                                break;
                            case "eventsAnd":
                                wfData.start.events = startBlock.data.formData.events
                                wfData.start.lifespan = startBlock.data.formData.lifespan
                                wfData.start.correlate = startBlock.data.formData.correlate
                                break;
                            case "eventsXor":
                                wfData.start.events = startBlock.data.formData.events
                                break;
                            default:
                                wfData.start.type = "default"
                                break;
                        }

                        // Find Start State
                        for (const outputID in startBlock.outputs) {
                            if (Object.hasOwnProperty.call(startBlock.outputs, outputID)) {
                                const output = startBlock.outputs[outputID];
                                if (output.connections.length === 0) {
                                    setError("Start Node is not connected to any node")
                                    return
                                }
                                startState = rawData[output.connections[0].node]
                                wfData.start.state = startState.data.id
                                break
                            }
                        }

                        // Set Transitions for main state flow
                        setConnections(startState.id, startBlock.id, null, rawData, wfData)
                        wfData.states.reverse()




                        // Create States for disconnected nodes
                        for (const nodeID in rawData) {
                            const rawNode = rawData[nodeID]

                            // Add primitive states with no input connections
                            if (!rawNode.compiled && rawNode.data.family === "primitive" && !nodeGetInputConnections(rawNode)) {
                                setConnections(rawNode.id, null, null, rawData, wfData)
                            }
                        }



                        if (updateWorkflow) {
                            wfData = {...unhandledData, ...wfData}
                            const workflowStr = unescapeJSStrings(PrettyYAML.stringify(wfData))
                            updateWorkflow(workflowStr)
                        } else {
                            console.warn("updateWorkflow callback missing")
                        }
                    }}>
                        <VscGear style={{ fontSize: "256px", width: "48px" }} />
                        <div>Compile</div>
                    </div>
                    <div className='toolbar-btn' onClick={() => {
                        if (actionDrawerMinWidth === 0) {
                            setActionDrawerMinWidth(MaxDrawerSize)
                            setActionDrawerWidth(actionDrawerWidthOld)

                            // Hide Functions
                            setFunctionDrawerMinWidth(0)
                            setFunctionDrawerWidth(0)
                        } else {
                            setActionDrawerMinWidth(0)
                            setActionDrawerWidth(0)
                        }
                    }}>
                        {actionDrawerMinWidth === 0 ?
                            <>
                                <VscListUnordered style={{ fontSize: "256px", width: "48px" }} />
                                <div>Show Nodes</div>
                            </>
                            :
                            <>
                                <VscListUnordered style={{ fontSize: "256px", width: "48px" }} />
                                <div>Hide Nodes</div>
                            </>
                        }
                    </div>
                    <div className='toolbar-btn' onClick={() => {
                        if (functionDrawerMinWidth === 0) {
                            setFunctionDrawerMinWidth(MaxDrawerSize)
                            setFunctionDrawerWidth(actionDrawerWidthOld)

                            // Hide Node Actions
                            setActionDrawerMinWidth(0)
                            setActionDrawerWidth(0)
                        } else {
                            setFunctionDrawerMinWidth(0)
                            setFunctionDrawerWidth(0)
                        }
                    }}>
                        {functionDrawerMinWidth === 0 ?
                            <>
                                <VscSymbolEvent style={{ fontSize: "256px", width: "48px" }} />
                                <div>Show Functions</div>
                            </>
                            :
                            <>
                                <VscSymbolEvent style={{ fontSize: "256px", width: "48px" }} />
                                <div>Hide Functions</div>
                            </>
                        }
                    </div>
                    <div className='toolbar-btn' onClick={() => {
                        sortNodes(diagramEditor)
                    }}>
                        <VscFileCode style={{ fontSize: "256px", width: "48px" }} />
                        <div>Format Nodes</div>
                    </div>
                </div>
                <FlexBox style={{ overflow: "hidden" }}>
                    <div
                        style={{
                            width: '100%',
                            display: 'flex',
                            overflow: 'hidden',
                            position: "relative"
                        }}
                    >
                        <Resizable
                            style={{ ...resizeStyle, pointerEvents: actionDrawerWidth === 0 ? "none" : "", visibility: actionDrawerWidth === 0 ? "hidden" : "visible" }}
                            size={{ width: actionDrawerWidth, height: "100%" }}
                            onResizeStop={(e, direction, ref, d) => {
                                setActionDrawerWidthOld(actionDrawerWidth + d.width)
                                setActionDrawerWidth(actionDrawerWidth + d.width)
                            }}
                            maxWidth="60%"
                            minWidth={actionDrawerMinWidth}
                        >
                            <div className={"panel left"} style={{ display: "flex" }}>
                                <div style={{ width: "100%", margin: "2px 0px 2px 4px" }}>
                                    <Actions />
                                </div>

                            </div>
                        </Resizable>
                        <Resizable
                            style={{ ...resizeStyle, pointerEvents: functionDrawerWidth === 0 ? "none" : "", visibility: functionDrawerWidth === 0 ? "hidden" : "visible" }}
                            size={{ width: functionDrawerWidth, height: "100%" }}
                            onResizeStop={(e, direction, ref, d) => {
                                setActionDrawerWidthOld(functionDrawerWidth + d.width)
                                setFunctionDrawerWidth(functionDrawerWidth + d.width)
                            }}
                            maxWidth="900"
                            minWidth={functionDrawerMinWidth}
                        >
                            <div className={"panel left"} style={{ display: "flex" }}>
                                <div style={{ width: "100%", margin: "2px 0px 2px 4px" }}>
                                    <FunctionsList functionDrawerWidth={functionDrawerWidth} functionList={functionList} setFunctionList={setFunctionList} namespace={namespace} />
                                </div>
                            </div>
                        </Resizable>
                        <div id="drawflow" style={{ height: "100%", width: "100%" }}
                            onDragOver={(ev) => {
                                ev.preventDefault();
                            }}
                            onDrop={(ev) => {
                                ev.preventDefault();
                                const nodeIndex = ev.dataTransfer.getData("nodeIndex");
                                const functionIndex = ev.dataTransfer.getData("functionIndex");
                                var newNode;

                                // Select NodeStateAction if function was dropped to quick create node
                                if (functionIndex !== "") {
                                    newNode = NodeStateAction
                                    newNode.data.formData = {
                                        action: {
                                            function: functionList[functionIndex].id
                                        }
                                    }
                                } else {
                                    newNode = ActionsNodes[nodeIndex]
                                }

                                setBlock(true)
                                CreateNode(diagramEditor, newNode, ev.clientX, ev.clientY)
                            }}
                            onContextMenu={(ev) => {
                                ev.preventDefault()
                                setContextMenuAnchorPoint({ x: ev.pageX, y: ev.pageY })

                                const parentIsNode = ev.target.offsetParent.className.includes("drawflow-node")
                                const targetIsNode = ev.target.className.includes("drawflow-node")

                                if (parentIsNode || targetIsNode) {
                                    // If Context menu event in on node div element
                                    setShowContextMenu(false)
                                    setShowNodeContextMenu(true)
                                } else {
                                    setShowContextMenu(true)
                                    setShowNodeContextMenu(false)
                                }

                                ev.stopPropagation()
                            }}
                        >
                        </div>
                        <ModalHeadless
                            visible={nodeDetailsVisible}
                            setVisible={setNodeDetailsVisible}
                            modalStyle={{ width: "60vw" }}
                            title={`Node Details: ${selectedNode ? selectedNode.data.id : ""}`}
                            actionButtons={[
                                ButtonDefinition("Submit", () => {
                                    setBlock(true)

                                    formRef.click()
                                    const ajv = new Ajv()
                                    const validate = ajv.compile(selectedNodeSchema)

                                    // Check if form data is valid
                                    if (!validate(selectedNodeFormData)) {
                                        throw Error("Invalid Values")
                                    }

                                    const updatedNode = {
                                        ...selectedNode,
                                        data: {
                                            ...selectedNode.data,
                                            formData: selectedNodeFormData,
                                            init: true
                                        }
                                    }

                                    // Preflight custom formData
                                    let onSubmitValidateCallback = onValidateSubmitCallbackMap[updatedNode.name]
                                    if (onSubmitValidateCallback) {
                                        onSubmitValidateCallback(selectedNodeFormData)
                                    } else {
                                        DefaultValidateSubmitCallbackMap(selectedNodeFormData)
                                    }


                                    // Update form data into node
                                    diagramEditor.updateNodeDataFromId(updatedNode.id, updatedNode.data)

                                    // Do Custom callback logic if it exists for data type
                                    let onSubmitCallback = onSubmitCallbackMap[updatedNode.name]
                                    if (onSubmitCallback) {
                                        onSubmitCallback(updatedNode.id, diagramEditor)
                                    } else {
                                        // Update SelectedNode state to updated state
                                        setSelectedNode(updatedNode)
                                    }

                                    // Track that node has data set
                                    setNodeInitTracker((old) => {
                                        old[selectedNode.id] = { init: updatedNode.data.init, stateID: updatedNode.data.id }
                                        return {
                                            ...old
                                        }
                                    })

                                    setOldSelectedNodeFormData(selectedNodeFormData)
                                }, "small blue", () => { }, true, false),
                                ButtonDefinition("Cancel", async () => {
                                    setSelectedNodeFormData(oldSelectedNodeFormData)
                                }, "small light", () => { }, true, false)
                            ]}
                        >
                            <Form
                                id={"builder-form"}
                                onSubmit={(form) => { }}
                                schema={selectedNodeSchema}
                                uiSchema={selectedNodeSchemaUI}
                                formData={selectedNodeFormData}
                                widgets={CustomWidgets}
                                onChange={(e) => {
                                    setSelectedNodeFormData(e.formData)
                                }}
                            >
                                <button ref={setFormRef} style={{ display: "none" }} />
                            </Form>
                        </ModalHeadless>
                        <ModalHeadless
                            visible={nodeIDModalVisible}
                            setVisible={setNodeIDModalVisible}
                            title={`Node Details: ${selectedNode ? selectedNode.data.id : ""}`}
                            modalStyle={{ width: "20vw" }}
                            actionButtons={[
                                ButtonDefinition("Save", () => {
                                    setError(null)
                                    // Update id to node data
                                    const updatedNode = {
                                        ...selectedNode,
                                        data: {
                                            ...selectedNode.data,
                                            formData: selectedNodeFormData
                                        }
                                    }

                                    updatedNode.data.id = newNodeID

                                    // Update form data into node
                                    diagramEditor.updateNodeDataFromId(updatedNode.id, updatedNode.data)

                                    setSelectedNode(updatedNode)

                                    // Track that node has data set
                                    setNodeInitTracker((old) => {
                                        old[selectedNode.id] = { init: updatedNode.data.init, stateID: updatedNode.data.id }
                                        return {
                                            ...old
                                        }
                                    })

                                    setNewNodeID("")
                                }, "small blue", () => { }, true, false),
                                ButtonDefinition("Cancel", async () => {
                                    setNewNodeID("")
                                }, "small light", () => { }, true, false)
                            ]}
                        >
                            <FlexBox className="col center" style={{ margin: "8px 16px 8px 16px" }}>
                                <FlexBox className="row center">
                                    <span style={{ whiteSpace: "nowrap", paddingRight: "6px" }}>
                                        Node ID:
                                    </span>
                                    <input type="text" className='nodeid-input' value={newNodeID} onChange={((e) => {
                                        setNewNodeID(e.target.value)
                                    })}>
                                    </input>
                                </FlexBox>
                            </FlexBox>
                        </ModalHeadless>
                    </div>
                </FlexBox>
                {/* </div> */}
            </FlexBox>
        </>
    )
}