import { ActionsNodes, NodeErrorBlock, NodeStartBlock } from "./nodes";
import YAML from 'js-yaml'
import prettyYAML from "json-to-pretty-yaml"
import { CreateNode, sortNodes, unescapeJSStrings } from "./util";

export function importFromYAML(diagramEditor, setFunctions, wfYAML) {
    const wfData = YAML.load(wfYAML)
    let nodeIDToStateIDMap = {}
    let catchNodes = []
    let pos = { x: 20, y: 200 }

    // Set functions
    if (wfData.functions) {
        setFunctions(wfData.functions)
    }

    // Add StartNode
    let startNode = JSON.parse(JSON.stringify((NodeStartBlock)))
    if (wfData.start) {
        startNode.data.formData = wfData.start
        startNode.data.init = true
    }

    const startNodeID = CreateNode(diagramEditor, startNode, pos.x, pos.y, true)

    // Iterate over states
    for (let i = 0; i < wfData.states.length; i++) {
        const state = wfData.states[i];
        const result = ActionsNodes.filter(ActionsNode => ActionsNode.family === "primitive" && ActionsNode.type === state.type)

        if (result.length === 0) {
            console.warn("State type not found when importing from YAML")
            continue
        }

        // Offset X
        pos.x += 220

        // Add Node to Diagram (DEEP COPY to avoid references changes)
        let newNode = JSON.parse(JSON.stringify(result[0]))

        // Create nodeData from state
        let newNodeData = Object.assign({}, state)

        delete newNodeData["type"]
        delete newNodeData["id"]
        delete newNodeData["catch"]
        // Convert transform

        const transformCallback = importProcessTransformCallback[newNode.name]
        if (transformCallback) {
            transformCallback(newNodeData)
        } else {
            const defaultCallback = importProcessTransformCallback["Default"]
            defaultCallback(newNodeData)
        }

        newNode.data.id = state.id
        newNode.data.formData = newNodeData
        newNode.data.init = true
        const nodeID = CreateNode(diagramEditor, newNode, pos.x, pos.y, true)

        // Add Catch Node
        if (state.catch) {
            pos.x += 220
            let errorNode = JSON.parse(JSON.stringify(NodeErrorBlock))
            const catchNodeRef = `SPECIAL-ERROR: ` + state.id
            errorNode.data.formData = state.catch
            errorNode.connections.output = state.catch.length
            errorNode.data.init = true
            const catchNodeID = CreateNode(diagramEditor, errorNode, pos.x, pos.y, true)
            nodeIDToStateIDMap[catchNodeRef] = catchNodeID

            catchNodes.push({ id: catchNodeID, catch: state.catch })
        }

        nodeIDToStateIDMap[state.id] = nodeID
    }


    // Connect Start Node to first state
    if (startNode.data.formData.state) {
        const firstNodeID = nodeIDToStateIDMap[startNode.data.formData.state]
        diagramEditor.addConnection(startNodeID, firstNodeID, 'output_1', 'input_1')
    } else {
        const firstNodeID = nodeIDToStateIDMap[wfData.states[0].id]
        diagramEditor.addConnection(startNodeID, firstNodeID, 'output_1', 'input_1')
    }

    // Iterate over states again and create connections
    for (let i = 0; i < wfData.states.length; i++) {
        const state = wfData.states[i];
        const nodeID = nodeIDToStateIDMap[state.id]
        const catchNodeRef = `SPECIAL-ERROR: ` + state.id
        const node = diagramEditor.getNodeFromId(nodeID)

        if (state.catch) {
            const catchNodeID = nodeIDToStateIDMap[catchNodeRef]
            diagramEditor.addConnection(nodeID, catchNodeID, 'output_2', 'input_1')
        }

        let connCallback = importConnectionsCallbackMap[node.name]
        if (connCallback) {
            connCallback(diagramEditor, state, nodeID, nodeIDToStateIDMap)
            continue
        }

        // Default Add Connections
        if (state.transition) {
            const nextNodeID = nodeIDToStateIDMap[state.transition]
            diagramEditor.addConnection(nodeID, nextNodeID, 'output_1', 'input_1')
        }
    }

    for (let i = 0; i < catchNodes.length; i++) {
        const catchNode = catchNodes[i]
        for (let j = 0; j < catchNode.catch.length; j++) {
            const err = catchNode.catch[j];
            const nextNodeID = nodeIDToStateIDMap[err.transition]
            diagramEditor.addConnection(catchNode.id, nextNodeID, `output_${j + 1}`, 'input_1')
        }
    }

    sortNodes(diagramEditor)

    // Preserve and return any other unhandled data
    delete wfData["functions"]
    delete wfData["states"]
    delete wfData["start"]

    return wfData
}

const importConnectionsCallbackMap = {
    "StateSwitch": (diagramEditor, state, nodeID, nodeIDToStateIDMap) => {
        if (state.defaultTransition) {
            const nextNodeID = nodeIDToStateIDMap[state.defaultTransition]
            diagramEditor.addConnection(nodeID, nextNodeID, 'output_1', 'input_1')
        }

        if (state.conditions) {
            for (let i = 0; i < state.conditions.length; i++) {
                const cond = state.conditions[i];
                const nextNodeID = nodeIDToStateIDMap[cond.transition]
                const nextNode = diagramEditor.getNodeFromId(nextNodeID)
                const newPosY = nextNode.pos_y + (100 * (i + 1));

                diagramEditor.drawflow.drawflow.Home.data[nextNodeID].pos_y = newPosY;
                document.getElementById(`node-${nextNodeID}`).style.top = `${newPosY}px`;
                diagramEditor.updateConnectionNodes(`node-${nextNodeID}`);

                diagramEditor.addNodeOutput(nodeID)
                diagramEditor.addConnection(nodeID, nextNodeID, `output_${i + 3}`, 'input_1')

            }
        }

    },
    "StateEventXor": (diagramEditor, state, nodeID, nodeIDToStateIDMap) => {
        if (state.events) {
            for (let i = 0; i < state.events.length; i++) {
                const event = state.events[i];
                const nextNodeID = nodeIDToStateIDMap[event.transition]
                const nextNode = diagramEditor.getNodeFromId(nextNodeID)
                const newPosY = nextNode.pos_y + (100 * (i + 1));

                diagramEditor.drawflow.drawflow.Home.data[nextNodeID].pos_y = newPosY;
                document.getElementById(`node-${nextNodeID}`).style.top = `${newPosY}px`;
                diagramEditor.updateConnectionNodes(`node-${nextNodeID}`);

                diagramEditor.addNodeOutput(nodeID)
                diagramEditor.addConnection(nodeID, nextNodeID, `output_${i + 2}`, 'input_1')

            }
        } else {
            console.warn("StateEventXor did not have events")
        }

    },
    // "StateGenerateEvent": (diagramEditor, states, nodeIDToStateIDMap) => {

    // },
    // "StateSetter": (diagramEditor, states, nodeIDToStateIDMap) => {

    // }
}

const objectDepth = (o) => Object(o) === o ? 1 + Math.max(-1, ...Object.values(o).map(objectDepth)) : 0

function importDefaultProcessTransformCallback(state, transformKey) {
    const oldTransform = state[transformKey]
    if (!oldTransform) {
        // No transform
        delete state[transformKey]
    } else if (typeof state[transformKey] === "object") {
        delete state[transformKey]

        // check depth
        const transformDepth = objectDepth(oldTransform)
        if (transformDepth > 1) {
            const yamlString = unescapeJSStrings(prettyYAML.stringify(oldTransform))
            state[transformKey] = {
                selectionType: "YAML",
                "rawYAML": yamlString
            }
        } else {
            state[transformKey] = {
                selectionType: "Key Value",
                "keyValue": oldTransform
            }
        }
    }  else if (oldTransform.trim().startsWith(`js(`) && oldTransform.trim().endsWith(")")){
        let javascriptString = ""
        let oldTransformArr = oldTransform.split("\n")
        for (let i = 1; i < oldTransformArr.length-1; i++) {
            const line = oldTransformArr[i].startsWith("  ") ? oldTransformArr[i].slice(2) : oldTransformArr[i];
            javascriptString += line + "\n"
        }

        state[transformKey] = {
            selectionType: "Javascript",
            "js": javascriptString
        }

    } else {
        state[transformKey] = {
            selectionType: "JQ Query",
            "jqQuery": oldTransform
        }
    }
}

function importConvertObjectToArray(state, objectKey) {
    const oldObject = state[objectKey]
    if (!oldObject) {
        delete state[objectKey]
    } else if (typeof state[objectKey] === "object") {
        delete state[objectKey]

        state[objectKey] = [{
            ...oldObject
        }]
    }
}

// TODO: This should be changed from transform callback to just general processing required form state to node
const importProcessTransformCallback = {
    "Default": (state) => { importDefaultProcessTransformCallback(state, "transform") },
    "StateEventXor": (state) => {
        for (let i = 0; i < state.events.length; i++) {
            importDefaultProcessTransformCallback(state.events[i], "transform")
        }
    },
    "StateGenerateEvent": (state) => {
        importDefaultProcessTransformCallback(state.event, "data")
        importDefaultProcessTransformCallback(state, "transform")
    },
    "StateSetter": (state) => {
        for (let i = 0; i < state.variables.length; i++) {
            importDefaultProcessTransformCallback(state.variables[i], "value")
        }

        importDefaultProcessTransformCallback(state, "transform")

    },
    "StateForeach": (state) => {
        importDefaultProcessTransformCallback(state.action, "input")
        importConvertObjectToArray(state.action, "retries")
        importDefaultProcessTransformCallback(state, "transform")
    },
    "StateAction": (state) => {
        importDefaultProcessTransformCallback(state.action, "input")
        importConvertObjectToArray(state.action, "retries")
        importDefaultProcessTransformCallback(state, "transform")
    },
    "StateSwitch": (state) => {
        importDefaultProcessTransformCallback(state, "defaultTransform")
        for (let i = 0; i < state.conditions.length; i++) {
            importDefaultProcessTransformCallback(state.conditions[i], "transform")
        }
    },
    "StateValidate": (state) => {
        // Convert schema to string
        state.schema = prettyYAML.stringify(state.schema ? state.schema : {})
    }
}




