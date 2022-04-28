import YAML from 'js-yaml'
import dagre from "dagre"


export function sortNodes(diagramEditor) {
    let allNodes = diagramEditor.export()

    // Create a new directed graph 
    var g = new dagre.graphlib.Graph({ordering: "out"});
    g.setGraph({});
    g.setDefaultEdgeLabel(function() { return {}; });
    g.graph().rankdir = "LR"

   for (const nodeKey in allNodes.drawflow.Home.data) {
       if (Object.hasOwnProperty.call(allNodes.drawflow.Home.data, nodeKey)) {
           const node = allNodes.drawflow.Home.data[nodeKey];
           g.setNode(`node_${nodeKey}`, { label: node.name,  nodeID: nodeKey, width: 144, height: 100});
       }
   }

   for (const nodeKey in allNodes.drawflow.Home.data) {
       if (Object.hasOwnProperty.call(allNodes.drawflow.Home.data, nodeKey)) {
           const node = allNodes.drawflow.Home.data[nodeKey];
           const outputs = nodeGetOutputConnections(node)
           if (outputs) {
               for (const key in outputs) {
                   for (const outputConnections of outputs[key]) {
                       g.setEdge(`node_${nodeKey}`, `node_${outputConnections.node}`);
                   }
               }
           }
       }
   }

   dagre.layout(g);

   g.nodes().forEach(function(v) {
       const gNode = g.node(v)
       allNodes.drawflow.Home.data[gNode.nodeID]["pos_x"] = gNode.x
       allNodes.drawflow.Home.data[gNode.nodeID]["pos_y"] = gNode.y
  });

  diagramEditor.import(allNodes)
}

export function nodeGetInputConnections(node) {
    let inputs = {}

    if (!node) {
        return undefined
    }

    if (!node.inputs) {
        return undefined
    }

    for (const inputKey of Object.keys(node.inputs)) {
        if (node.inputs[inputKey].connections.length > 0) {
            inputs[inputKey] = node.inputs[inputKey].connections
        }
    }

    if (Object.keys(inputs).length === 0) {
        return undefined
    }

    return inputs
}

export function nodeGetOutputConnections(node) {
    let outputs = {}

    if (!node) {
        return undefined
    }

    if (!node.outputs) {
        return undefined
    }

    for (const outputKey of Object.keys(node.outputs)) {
        if (node.outputs[outputKey].connections.length > 0) {
            outputs[outputKey] = node.outputs[outputKey].connections
        }
    }

    if (Object.keys(outputs).length === 0) {
        return undefined
    }

    return outputs
}

export function insertHangingNodes(nodeID, previousNodeID, previousState, rawData, wfData) {
    // Use Custom connection callback logic if it exists for data type
    rawData[nodeID].compiled = true
    let currentNode = rawData[nodeID]
    let connCallback = connectionsCallbackMap[currentNode.name]
    if (connCallback) {
        connCallback(nodeID, previousNodeID, previousState, rawData, wfData)
        return
    }



    let state = { id: currentNode.data.id, type: currentNode.data.type, ...currentNode.data.formData }
    processTransform(state, "transform")

    // Default connections logic
    processConnection(nodeID, rawData, state, wfData)

    return
}

//  processTransform : Converts json schema form to direktiv yaml on the properties that accept jq.
//  Because we support both a jq string a key value json schema input, we need to process these values into
//  something direkitv yaml can use. 
//
//  stateData: The parent of a value that contains a jqQuery or key value property
//  transformKey: The key of the property that contains this value. This will usually be 'transform'
//  but there are some scenarios where it is something else. e.g. "data" from generateEvent
function processTransform(stateData, transformKey) {
    if (!stateData || !stateData[transformKey]) {
        return
    }

    const selectionType = stateData[transformKey]["selectionType"]
    const keyValue = stateData[transformKey]["keyValue"] ? stateData[transformKey]["keyValue"] : {}
    const jqQuery = stateData[transformKey]["jqQuery"] ? stateData[transformKey]["jqQuery"] : ""
    const rawYAML = stateData[transformKey]["rawYAML"] ? stateData[transformKey]["rawYAML"] : ""
    const js = stateData[transformKey]["js"] ? stateData[transformKey]["js"] : ""

    delete stateData[transformKey]["keyValue"]
    delete stateData[transformKey]["jqQuery"]
    delete stateData[transformKey]["selectionType"]
    delete stateData[transformKey]["rawYAML"]
    delete stateData[transformKey]["js"]

    if (selectionType && selectionType === "Key Value") {
        stateData[transformKey] = { ...keyValue }
    } else if (selectionType && selectionType === "YAML") {
        stateData[transformKey] = YAML.load(rawYAML)
    } else if (selectionType && selectionType === "JQ Query") {
        stateData[transformKey] = jqQuery
    } else if (selectionType && selectionType === "Javascript") {
        stateData[transformKey] = `js(\n${js}\n)`
    }

    if (stateData[transformKey] === "" || stateData[transformKey] === {}) {
        delete stateData[transformKey]
    }

    return
}

function processArrayToObject(stateData, objectKey) {
    if (!stateData || !stateData[objectKey]) {
        return
    }

    const oldArray = stateData[objectKey]
    if (!oldArray || !Array.isArray(oldArray) || oldArray.length <= 0) {
        delete stateData[objectKey]
    } else if (Array.isArray(oldArray)) {
        delete stateData[objectKey]

        stateData[objectKey] = { ...oldArray[0] }
    }

    return
}

// Recursively walk through nodes and sets transitions of each state
export function setConnections(nodeID, previousNodeID, previousState, rawData, wfData) {

    // Stop recursive walk node has already been compiled
    // If we dont do this we'll there is a chance that we create the same state multiple times
    if (rawData[nodeID].compiled) {
        return
    }

    // Use Custom connection callback logic if it exists for data type
    rawData[nodeID].compiled = true
    let currentNode = rawData[nodeID]
    let connCallback = connectionsCallbackMap[currentNode.name]
    if (connCallback) {
        connCallback(nodeID, previousNodeID, previousState, rawData, wfData)
        return
    }



    let state = { id: currentNode.data.id, type: currentNode.data.type, ...currentNode.data.formData }
    processTransform(state, "transform")

    // Default connections logic
    processConnection(nodeID, rawData, state, wfData)

    return
}

// default handler for connections
function processConnection(nodeID, rawData, state, wfData) {
    const currentNode = rawData[nodeID]
    const outputKeys = Object.keys(rawData[nodeID].outputs)

    // Default connections logic
    for (let i = 0; i < outputKeys.length; i++) {
        const outputID = outputKeys[i];

        if (Object.hasOwnProperty.call(currentNode.outputs, outputID)) {
            const output = currentNode.outputs[outputID];
            if (output.connections.length > 0) {
                const nextNode = rawData[output.connections[0].node]

                // Only use first node output connection for transition
                if (i === 0) {
                    state.transition = nextNode.data.id
                }

                setConnections(output.connections[0].node, nodeID, state, rawData, wfData)
            }
        }
    }

    wfData.states.push(state)
}

// connectionsCallbackMap : Map of functions to be used in setConnections function
const connectionsCallbackMap = {
    "CatchError": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let stateCatch = rawData[nodeID].data.formData
        const outputKeys = Object.keys(rawData[nodeID].outputs)

        // Add transitions to catched errors if connections are set
        for (let i = 0; i < outputKeys.length; i++) {
            const outputID = outputKeys[i];

            if (Object.hasOwnProperty.call(rawData[nodeID].outputs, outputID)) {
                const output = rawData[nodeID].outputs[outputID];
                if (output.connections.length > 0) {
                    const nextNode = rawData[output.connections[0].node]
                    stateCatch[i].transition = nextNode.data.id
                }
            }

        }

        // Only set non-empty catches
        if (Object.keys(stateCatch).length > 0) {
            previousState.catch = stateCatch
        }

        for (let i = 0; i < outputKeys.length; i++) {
            const outputID = outputKeys[i];
            if (Object.hasOwnProperty.call(rawData[nodeID].outputs, outputID)) {
                const output = rawData[nodeID].outputs[outputID];

                if (output.connections.length > 0) {
                    setConnections(output.connections[0].node, nodeID, previousState, rawData, wfData)
                }
            }

        }


    },
    "StateSwitch": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }
        processTransform(state, "defaultTransform")

        const outputKeys = Object.keys(rawData[nodeID].outputs)
        for (let i = 0; i < outputKeys.length; i++) {
            const outputID = outputKeys[i];

            if (Object.hasOwnProperty.call(rawData[nodeID].outputs, outputID)) {
                const output = rawData[nodeID].outputs[outputID];
                if (output.connections.length > 0) {
                    const nextNode = rawData[output.connections[0].node]

                    if (i === 0) {
                        // First Node Connection
                        state.defaultTransition = nextNode.data.id
                    } else if (i > 1) {
                        // Skip Second node connection because of error catcher
                        state.conditions[i - 2].transition = nextNode.data.id
                    }

                    processTransform(state.conditions[i], "transform")

                    setConnections(output.connections[0].node, nodeID, state, rawData, wfData)
                }
            }

        }

        wfData.states.push(state)
    },
    "StateEventXor": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }

        for (let i = 0; i < state.events.length; i++) {
            processTransform(state.events[i], "transform")
        }

        const outputKeys = Object.keys(rawData[nodeID].outputs)
        for (let i = 0; i < outputKeys.length; i++) {
            const outputID = outputKeys[i];

            if (Object.hasOwnProperty.call(rawData[nodeID].outputs, outputID)) {
                const nodeOutput = rawData[nodeID].outputs[outputID];
                if (nodeOutput.connections.length > 0) {
                    // skip first node connection (error catcher)
                    if (i > 0) {
                        const nextNode = rawData[nodeOutput.connections[0].node]
                        state.events[i-1].transition = nextNode.data.id
                    }

                    setConnections(nodeOutput.connections[0].node, nodeID, state, rawData, wfData)
                }
            }

        }

        wfData.states.push(state)
    },
    "StateGenerateEvent": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }

        processTransform(state.event, "data")
        processTransform(state, "transform")

        // Default connections logic
        processConnection(nodeID, rawData, state, wfData)
    },
    "StateSetter": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }

        for (let i = 0; i < state.variables.length; i++) {
            processTransform(state.variables[i], "value")
        }

        processTransform(state, "transform")

        // Default connections logic
        processConnection(nodeID, rawData, state, wfData)
    },
    "StateAction": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }

        processArrayToObject(state.action, "retries")
        processTransform(state.action, "input")
        processTransform(state, "transform")

        // Default connections logic
        processConnection(nodeID, rawData, state, wfData)
    },
    "StateForeach": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }

        processArrayToObject(state.action, "retries")
        processTransform(state.action, "input")
        processTransform(state, "transform")

        // Default connections logic
        processConnection(nodeID, rawData, state, wfData)
    },
    "StateValidate": (nodeID, previousNodeID, previousState, rawData, wfData) => {
        let state = { id: rawData[nodeID].data.id, type: rawData[nodeID].data.type, ...rawData[nodeID].data.formData }
        state.schema = YAML.load(state.schema)

        // Default connections logic
        processConnection(nodeID, rawData, state, wfData)
    },
}

export function DefaultValidateSubmitCallbackMap(formData) {
    validateFormTransform(formData.transform)
}

function validateFormTransform(formTransformData) {
    if (!formTransformData) {
        return
    }

    if (formTransformData.selectionType && formTransformData.selectionType === "YAML" && formTransformData.rawYAML !== "") {
        try {
            YAML.load(formTransformData.rawYAML)
        } catch (e) {
            throw Error(`Invalid Raw YAML: ${e.reason}`)
        }
    }
}

export const onValidateSubmitCallbackMap = {
    "Default": DefaultValidateSubmitCallbackMap,
    "StateEventXor": (formData) => {
        for (let i = 0; i < formData.events.length; i++) {
            validateFormTransform(formData.events[i].transform)
        }
    },
    "StateGenerateEvent": (formData) => {
        validateFormTransform(formData.event.data)
        validateFormTransform(formData.transform)
    },
    "StateSetter": (formData) => {
        for (let i = 0; i < formData.variables.length; i++) {
            validateFormTransform(formData.variables[i].value)
        }
        validateFormTransform(formData.transform)
    },
    "StateAction": (formData) => {
        validateFormTransform(formData.transform)
        validateFormTransform(formData.action.input)
    },
    "StateForeach": (formData) => {
        validateFormTransform(formData.transform)
        validateFormTransform(formData.action.input)
    },
    "StateValidate": (formData) => {
        try {
            YAML.load(formData.schema)
        } catch (e) {
            throw Error(`Validate Schema Invalid: ${e.reason}`)
        }
    },
}

export const onSubmitCallbackMap = {
    "StateSwitch": (nodeID, diagramEditor) => {
        const node = diagramEditor.getNodeFromId(nodeID)
        let conditionsLength = node.data.formData.conditions ? node.data.formData.conditions.length : 0
        // outputLen : Is the outputs minus the error and default transition outputs
        const outputLen = Object.keys(node.outputs).length - 2

        // Add Missing Node Outputs
        for (let i = outputLen; i < conditionsLength; i++) {
            diagramEditor.addNodeOutput(node.id)
        }

        // Remove excess node outputs
        for (let i = conditionsLength; i < outputLen; i++) {
            diagramEditor.removeNodeOutput(node.id, `output_${i + 2}`)
        }

    },
    "StateEventXor": (nodeID, diagramEditor) => {
        const node = diagramEditor.getNodeFromId(nodeID)
        let eventsLength = node.data.formData.events ? node.data.formData.events.length : 0
        const outputLen = Object.keys(node.outputs).length - 1

        // Add Missing Node Outputs
        for (let i = outputLen; i < eventsLength; i++) {
            diagramEditor.addNodeOutput(node.id)
        }

        // Remove excess node outputs
        for (let i = eventsLength; i < outputLen; i++) {
            diagramEditor.removeNodeOutput(node.id, `output_${i + 1}`)
        }

    },
    "CatchError": (nodeID, diagramEditor) => {
        const node = diagramEditor.getNodeFromId(nodeID)
        let errorsLength = node.data.formData ? node.data.formData.length : 0
        const outputLen = Object.keys(node.outputs).length

        // Add Missing Node Outputs
        for (let i = outputLen; i < errorsLength; i++) {
            diagramEditor.addNodeOutput(node.id)
        }

        // Remove excess node outputs
        for (let i = errorsLength; i < outputLen; i++) {
            diagramEditor.removeNodeOutput(node.id, `output_${i}`)
        }

    }
}


export function CreateNode(diagramEditor, node, clientX, clientY, rawXY) {
    var newNodeHTML
    let posX = clientX
    let posY = clientY

    // Optional Coordinates processing.
    if (!rawXY) {
        posX = clientX * (diagramEditor.precanvas.clientWidth / (diagramEditor.precanvas.clientWidth * diagramEditor.zoom)) - (diagramEditor.precanvas.getBoundingClientRect().x * (diagramEditor.precanvas.clientWidth / (diagramEditor.precanvas.clientWidth * diagramEditor.zoom)));
        posY = clientY * (diagramEditor.precanvas.clientHeight / (diagramEditor.precanvas.clientHeight * diagramEditor.zoom)) - (diagramEditor.precanvas.getBoundingClientRect().y * (diagramEditor.precanvas.clientHeight / (diagramEditor.precanvas.clientHeight * diagramEditor.zoom)));

    }

    // Generate HTML
    switch (node.family) {
        case "special":
            newNodeHTML = `<div class="node-labels">
            <div>
                <span class="label-type">${node.html}</span>
            </div>
        </div>`
            break;
        case "primitive":
            newNodeHTML = `<div class="node-labels">
            <div>
                ID: <input class="label-id" type="text" df-id>
            </div>
            <div>
                Type: <span class="label-type">${node.html}</span>
            </div>
        </div>`
            break
        default:
            newNodeHTML = `<div class="node-labels">
            <div>
                <span class="label-type">${node.html}</span>
            </div>
        </div>`
            break;
    }

    // Add Action to HTML
    if (node.info.actions) {
        newNodeHTML += `
    <div class="node-actions">
        <span id="node-btn" class="node-btn">
            ...
        </span>
    </div>`
    }


    return diagramEditor.addNode(node.name, node.connections.input, node.connections.output, posX, posY, `node ${node.family} type-${node.type}`, { family: node.family, type: node.type, init: !node.info.requiresInit, ...node.data }, newNodeHTML, false)
}

//  unescapeJSStrings : Accepts a line string that has an escaped js string value and then unescapes the string and converts it to a multi-line YAML string.
//  Example: 
//      Input: 
//        transform: "js(var ret = new Array();\nb = data['ping'].concat(data['resolve']);)"
//      Output:
//        transform: |
//          js(
//            var ret = new Array();
//            b = data['ping'].concat(data['resolve']);
//          )
export function unescapeJSStrings(str) {
    const symbol = `"js(`
    let newWorkflowStr = ""

    str.split("\n").forEach(lineStr => {
        const replaceIndex = lineStr.indexOf(symbol)
        if (replaceIndex > 0) {
            const jsStr = lineStr.slice(replaceIndex+1, -2).trim()

            // Count leading whitespaces
            let whiteSpaceCount = 0
            for (;  lineStr[whiteSpaceCount] === " "; whiteSpaceCount++) {}

            // Split javascript into multiline YAML string
            newWorkflowStr += lineStr.slice(0, replaceIndex-1) + " |\n"
            jsStr.split("\\n").forEach(jsLineStr => {
                const test = jsLineStr.replaceAll(`\\`, ``)
                newWorkflowStr += " ".repeat(whiteSpaceCount)+ "    " + test + "\n"
            });
        } else {
            newWorkflowStr += lineStr + "\n"
        }
    });

    return newWorkflowStr
}