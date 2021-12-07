import dagre from 'dagre'
import './style.css';
import ReactFlow, {ReactFlowProvider, MiniMap, isNode, Handle, useZoomPanHelper} from 'react-flow-renderer'
import React, { useEffect, useState } from "react"
import { position } from '../diagram'
import FlexBox from '../flexbox'
import {BiDotsVerticalRounded} from 'react-icons/bi';


export default function DependencyDiagram(props) {

    const {dependencies, type, workflow} = props
    console.log(dependencies)
    const [load, setLoad] = useState(true)
    const [elements, setElements] = useState([])

    useEffect(()=>{
        async function drawGraph() {
            if(load && type === "workflow" && dependencies !== null){
                const dagreGraph = new dagre.graphlib.Graph()
                dagreGraph.setDefaultEdgeLabel(() => ({}))
            
                const getLayoutedElements = (incomingEles, direction = 'TB') => {
                    const isHorizontal = direction === 'LR'
                    dagreGraph.setGraph({rankdir: 'lr'})
                
                    incomingEles.forEach((el)=>{
                        if(isNode(el)){
                            dagreGraph.setNode(el.id, {width: 100, height:40})
                        }else {
                            dagreGraph.setEdge(el.source, el.target, {width: 60, height: 60})
                        }
                    })
                
                    dagre.layout(dagreGraph)
                    
                    return incomingEles.map((el)=>{
                        if(isNode(el)){
                            const nodeWithPosition = dagreGraph.node(el.id)
                            el.targetPosition = isHorizontal ? 'left' : 'top'
                            el.sourcePosition = isHorizontal ? 'right' : 'bottom'
                
                            //hack to trigger refresh
                            el.position = {
                                x: nodeWithPosition.x + Math.random()/1000,
                                y: nodeWithPosition.y,
                            }
                        }
                        return el
                    })
                }

                let saveElements = generateElements(getLayoutedElements, dependencies, workflow)
                if(saveElements !== null) {
                    setElements(saveElements)
                }
                setLoad(false)
            }    
        }
        drawGraph()
    },[load, dependencies])


    if(load){
        return ""
    }

    if(type === "workflow") {
        return(
            <ReactFlowProvider>
                <DependencyGraph elements={elements}/>
            </ReactFlowProvider>
        )
    }

    return (
        <div>type provided not supported</div>
    )
}

function Found(props) {
    const {data} = props
    const {label, type} = data

    let includeActions = false;
    let actions = [];
    if (type.toLowerCase() === "subflow") {
        includeActions = true;
        actions.push(
            <BiDotsVerticalRounded key="state-action-1" /> // TODO - real action btn
        )
    }

    return(
        <div className="dependency-diagram state green-state" title={label}>
            <FlexBox className="state-info flex-centered col">
                <Handle
                    type="target"
                    position="left"
                    id="default"
                />
                <div className="state-type">
                    {type}
                </div>
                <div className="state-label">{label}</div>
                <Handle
                    type="source"
                    position="right"
                    id="default"
                /> 
            </FlexBox>
            { includeActions ? 
            <div className="state-actions">
                {actions}
            </div>:<></>}
        </div>
    )
}


function Missing(props) {
    const {data} = props
    const {label, type} = data

    return(
        <div className="dependency-diagram state red-state" title="This dependency does not exist.">
            <FlexBox className="state-info flex-centered col">
                <Handle
                    type="target"
                    position="left"
                    id="default"
                />
                <div className="state-type">
                    {type}
                </div>
                <div>{label}</div>
                <Handle
                    type="source"
                    position="right"
                    id="default"
                /> 
            </FlexBox>
        </div>
    )
}

function DependencyGraph(props) {
    const {elements} = props
    const { fitView } = useZoomPanHelper()

    useEffect(()=>{
        fitView()
    },[fitView])
    return (
        <ReactFlow
            elements={elements}
            nodeTypes={{
                missing: Missing,
                found: Found
            }}
            nodesDraggable={false}
            nodesConnectable={false}
            // elementsSelectable={false}
            paneMoveable={false}
        >
            <MiniMap
                    nodeColor={()=>{
                        return '#4497f5'
                    }}
            />
        </ReactFlow>
    )
}

function generateElements(getLayoutedElements, dependencies, workflow) {
    let newElements = []

    let parentNodes = Object.keys(dependencies.parents).map((obj)=>{
        return(
            {
                id: `parent-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'parent'
                },
                type: dependencies.parents[obj] ? "found": "missing"
            }
        )
    })

    let parentEdges = Object.keys(dependencies.parents).map((obj)=>{
        return(
            {
                id: `${obj}-${workflow}`,
                source: `parent-${obj}`,
                target: workflow,
                type: 'pathfinding',
                arrowHeadType: 'arrow'
            }
        )
    })

    newElements.push({
        id: workflow,
        position: position,
        data: {
            label: workflow,
            type: 'workflow'.toUpperCase()
        },
        type: 'found'
    })

    let secretNodes = Object.keys(dependencies.secrets).map((obj)=>{
        return(
            {
                id: `secret-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'secret'.toUpperCase()
                },
                type: dependencies.secrets[obj] ? "found": "missing"
            }
        )
    })
    let secretEdges = Object.keys(dependencies.secrets).map((obj)=>{
        return( {
            id: `${workflow}-secrets-${obj}`,
            source: workflow,
            target: `secret-${obj}`,
            type: 'pathfinding'
        })
    })

    let subflowNodes = Object.keys(dependencies.subflows).map((obj) => {
        return (
            {
                id: `subflow-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'subflow'.toUpperCase()
                },
                type: dependencies.subflows[obj] ? "found" : "missing"
            }
        )
    })
    let subflowEdges = Object.keys(dependencies.subflows).map((obj) => {
        return ({
            id: `${workflow}-subflows-${obj}`,
            source: workflow,
            target: `subflow-${obj}`,
            type: 'pathfinding'
        })
    })

    let globalNodes = Object.keys(dependencies.global_functions).map((obj)=>{
        return(
            {
                id: `g-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'knative-global'.toUpperCase()
                },
                type: dependencies.global_functions[obj] ? "found": "missing"
            }
        )
    })
    let globalEdges = Object.keys(dependencies.global_functions).map((obj)=>{
        return( {
            id: `${workflow}-g-${obj}`,
            source: workflow,
            target: `g-${obj}`,
            type: 'pathfinding',
            arrowHeadType: 'arrow'
        })
    })

    let namespaceNodes = Object.keys(dependencies.namespace_functions).map((obj)=>{
        return(
            {
                id: `ns-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'knative-namespace'.toUpperCase()
                },
                type: dependencies.namespace_functions[obj] ? "found": "missing"
            }
        )
    })
    let namespaceEdges = Object.keys(dependencies.namespace_functions).map((obj)=>{
        return( {
            id: `${workflow}-ns-${obj}`,
            source: workflow,
            target: `ns-${obj}`,
            type: 'pathfinding',
            arrowHeadType: 'arrow'
        })
    })

    let namespaceVarNodes = Object.keys(dependencies.namespace_variables).map((obj)=>{
        return(
            {
                id: `nsvar-${obj}`,
                position: position,
                data: {
                    label: obj,
                    type: 'namespace-variable'.toUpperCase()
                },
                type: dependencies.namespace_variables[obj] ? "found": "missing"
            }
        )
    })
    let namespaceVarEdges = Object.keys(dependencies.namespace_variables).map((obj)=>{
        return( {
            id: `${workflow}-nsvar-${obj}`,
            source: workflow,
            target: `nsvar-${obj}`,
            type: 'pathfinding',
            arrowHeadType: 'arrow'
        })
    })

    return getLayoutedElements(
        newElements.concat(
            parentNodes, parentEdges,
            secretNodes, secretEdges,
            globalNodes, globalEdges,
            namespaceNodes, namespaceEdges,
            subflowNodes, subflowEdges,
            namespaceVarNodes, namespaceVarEdges
        )
    )
}