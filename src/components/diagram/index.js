import dagre from 'dagre'
import ReactFlow, {ReactFlowProvider, MiniMap, isNode, Handle, useZoomPanHelper} from 'react-flow-renderer'
import React, { useEffect, useState } from "react"

import './style.css'
export const position = { x: 0, y: 0}


    // initialize the dagre graph
    const dagreGraph = new dagre.graphlib.Graph()
    dagreGraph.setDefaultEdgeLabel(() => ({}))
export default function WorkflowDiagram(props) {
    const {workflow, flow, instanceStatus, disabled} = props

    const [load, setLoad] = useState(true)
    const [elements, setElements] = useState([])
    const [ostatus, setOStatus] = useState(instanceStatus)


    useEffect(()=>{
    
                   
        const getLayoutedElements = (incomingEles, direction = 'TB') => {
            const isHorizontal = direction === 'LR'
            dagreGraph.setGraph({rankdir: 'lr'})
        
            incomingEles.forEach((el)=>{
                if(isNode(el)){
                    if(el.id === "startNode"|| el.id === "endNode"){
                        dagreGraph.setNode(el.id, {width: 40, height:40})
                    } else {
                        dagreGraph.setNode(el.id, {width: 100, height:40})
                    }
                } else {
                    if(el.source === "startNode") {
                        dagreGraph.setEdge(el.source, el.target, {width: 0, height: 20})
                    } else if(el.source === "endNode"){
                        dagreGraph.setEdge(el.source, el.target, {width: 30, height: 20})
                    } else {
                        dagreGraph.setEdge(el.source, el.target, {width: 60, height: 60})
                    }
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

        if(load && (workflow !== null || instanceStatus !== ostatus)) {
            let saveElements = generateElements(getLayoutedElements, workflow, flow, instanceStatus)
            if(saveElements !== null) {
                setElements(saveElements)
            }
            setOStatus(instanceStatus)
            setLoad(false)
        }

        // if status changes make sure to redraw
        if(instanceStatus !== ostatus) {
            let saveElements = generateElements(getLayoutedElements, workflow, flow, instanceStatus)
            if(saveElements !== null) {
                setElements(saveElements)
            }
            setOStatus(instanceStatus)
        }
        
    },[load, workflow,flow, instanceStatus, ostatus])

    if(load) {
        return <></>
    }

    return(
       <ReactFlowProvider>
           <ZoomPanDiagram disabled={disabled} elements={elements}/>
       </ReactFlowProvider>
    )
}

function ZoomPanDiagram(props) {
    const {elements, disabled} = props
    const { fitView } = useZoomPanHelper();

    useEffect(()=>{
        fitView()
    },[fitView])
    return (
        <ReactFlow
            elements={elements}
            nodeTypes={{
                state: State,
                start: Start,
                end: End
            }}
            nodesDraggable={disabled}
            nodesConnectable={disabled}
            elementsSelectable={disabled}
            paneMoveable={disabled}
        >
            <MiniMap
                    nodeColor={()=>{
                        return '#4497f5'
                    }}
            />
        </ReactFlow>
    )
}

function State(props) {
    const {data} = props
    const {label, type} = data
    return(
        <div className="state" style={{width:"80px", height:"30px"}}>
            <Handle
                type="target"
                position="left"
                id="default"
            />
            <div style={{display:"flex", padding:"1px", gap:"3px", alignItems:"center", fontSize:"6pt", textAlign:"left", borderBottom: "solid 1px rgba(0, 0, 0, 0.1)"}}> 
                <div style={{flex:"auto", fontWeight:"bold"}}>
                    {type}
                </div>
            </div>
            <h1 style={{fontWeight:"300", fontSize:"7pt", marginTop:"2px"}}>{label}</h1>
            <Handle
                type="source"
                position="right"
                id="default"
            /> 
        </div>
    )
}

function Start() {
    return(
        <div className="normal">
            <Handle
                type="source"
                position="right"
            />
            <div className="start" />
        </div>
    )
}

function End() {
    return(
        <div className="normal">
            <div className="end" />
             <Handle
                type="target"
                position="left"
            />
        </div>
    )
}

function generateElements(getLayoutedElements, value, flow, status) {
    let newElements = []

    if(value.states) {
        for(let i=0; i < value.states.length; i++) {
                let transitions = false
                // check if starting element
                if (i === 0) {
                    // starting element so create an edge to the state
                    if (value.start && value.start.state){
                        newElements.push({
                            id: `startNode-${value.start.state}`,
                            source: 'startNode',
                            target: value.start.state,
                            type: 'pathfinding',
                            arrowHeadType: 'arrow'
                        })
                    } else {
                        newElements.push({
                            id: `startNode-${value.states[i].id}`,
                            source: 'startNode',
                            target: value.states[i].id,
                            type: 'pathfinding',
                            arrowHeadType: 'arrow'
                        })
                    }

            
                }

                // push new state
                newElements.push({
                    id: value.states[i].id,
                    position: position,
                    data: {label: value.states[i].id, type: value.states[i].type, state: value.states[i], functions: value.functions},
                    type: 'state'
                })

                // check if the state has events
                if (value.states[i].events) {
                    for(let j=0; j < value.states[i].events.length; j++) {
                        if(value.states[i].events[j].transition) {
                            transitions = true
                            newElements.push({
                                id: `${value.states[i].id}-${value.states[i].events[j].transition}`,
                                source: value.states[i].id,
                                target: value.states[i].events[j].transition,
                                animated: false,
                                type: 'pathfinding',
                                arrowHeadType: 'arrow'
                            })
                        }
                    }
                }

                // Check if the state has conditions
                if(value.states[i].conditions) {
                    for(let y=0; y < value.states[i].conditions.length; y++) {
                        if(value.states[i].conditions[y].transition) {
                            newElements.push({
                                id: `${value.states[i].id}-${value.states[i].conditions[y].transition}`,
                                source: value.states[i].id,
                                target: value.states[i].conditions[y].transition,
                                animated: false,
                                type: 'pathfinding',
                                arrowHeadType: 'arrow'
                            })
                            transitions = true

                        }
                    }
                }

                // Check if state is catching things to transition to
                if(value.states[i].catch) {
                    for(let x=0; x < value.states[i].catch.length; x++) {
                        if(value.states[i].catch[x].transition) {
                            transitions = true

                            newElements.push({
                                id: `${value.states[i].id}-${value.states[i].catch[x].transition}`,
                                source: value.states[i].id,
                                target: value.states[i].catch[x].transition,
                                animated: false,
                                type: 'pathfinding',
                                arrowHeadType: 'arrow'
                            })
                        }
                    }
                }

                // check if transition and create edge to hit new state
                if(value.states[i].transition) {
                    transitions = true

                    newElements.push({
                        id: `${value.states[i].id}-${value.states[i].transition}`,
                        source: value.states[i].id,
                        target: value.states[i].transition,
                        animated: false,
                        type: 'pathfinding',
                        arrowHeadType: 'arrow'
                    })
                } else if(value.states[i].defaultTransition) {
                    transitions = true

                    newElements.push({
                        id: `${value.states[i].id}-${value.states[i].defaultTransition}`,
                        source: value.states[i].id,
                        target: value.states[i].defaultTransition,
                        animated: false,
                        type: 'pathfinding',
                        arrowHeadType: 'arrow'
                    })
                } else {
                        transitions = true
                        newElements.push({
                            id: `${value.states[i].id}-endNode`,
                            source: value.states[i].id,
                            target: `endNode`,
                            animated: false,
                            type: 'pathfinding',
                            arrowHeadType: 'arrow'
                        })
                }

                if(!transitions) {
                    // no transition add end state
                    newElements.push({
                        id: `${value.states[i].id}-endNode`,
                        source: value.states[i].id,
                        target: `endNode`,
                        type: 'pathfinding',
                        arrowHeadType: 'arrow'
                    })
                }
            }

            // push start node
            newElements.push({
                id: 'startNode',
                position: position,
                data: {label: ""},
                type: 'start',
                sourcePosition: 'right',
            })

            // push end node
            newElements.push({
                id:'endNode',
                type: 'end',
                data: {label: ""},
                position: position,
            })

            // Check flow array change edges to green if it passed 
            if(flow){
                // check flow for transitions
                for(let i=0; i < flow.length; i++) {
                    let noTransition = false
                    for(let j=0; j < newElements.length; j++) {
                        
                        // handle start node
                        if(newElements[j].source === "startNode" && newElements[j].target === flow[i]){
                            newElements[j].animated = true
                        }

                        
                        if(newElements[j].target === flow[i] && newElements[j].source === flow[i-1]) {
                            newElements[j].animated = true
                        } else if(newElements[j].id === flow[i]) {
                            if(!newElements[j].data.state.transition || !newElements[j].data.state.defaultTransition ){
                                noTransition = true
                            
                                if(newElements[j].data.state.catch) {
                                    for(let y=0; y < newElements[j].data.state.catch.length; y++) {
                                        if(newElements[j].data.state.catch[y].transition){
                                            noTransition = false
                                            if (newElements[j].data.label === flow[flow.length-1]) {
                                                noTransition = true
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }

                    if(noTransition) {
                        // transition to end state
                        // check if theres more flow if not its the end node
                        if(!flow[i+1]){
                            for(let j=0; j < newElements.length; j++) {
                                if(newElements[j].source === flow[i] && newElements[j].target === "endNode" && (status === "complete"|| status === "failed") ){
                                    newElements[j].animated = true
                                }
                            }
                        }
                    }
                }
        }
    }
    return getLayoutedElements(newElements)
}