import dagre from 'dagre'
import ReactFlow, {ReactFlowProvider, MiniMap, isNode, Handle, useZoomPanHelper} from 'react-flow-renderer'
import React, { useEffect, useState } from "react"
import { position } from '../diagram'


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
    return(
        <div className="state" style={{color:"white",width:"80px", height:"50px", backgroundColor:"#55BA86"}}>
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


function Missing(props) {
    const {data} = props
    const {label, type} = data
    return(
        <div title="This dependency does not exist." className="state" style={{color:"white", width:"80px", height:"50px", backgroundColor: "#EC4F79"}}>
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
            elementsSelectable={false}
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

    newElements.push({
        id: workflow,
        position: position,
        data: {
            label: workflow,
            type: 'workflow'
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
                    type: 'secret'
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
            type: 'pathfinding',
            arrowHeadType: 'arrow'
        })
    })

    return getLayoutedElements(newElements.concat(secretNodes, secretEdges))
}