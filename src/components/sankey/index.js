import {useState, useEffect} from 'react'
import './style.css';
import AutoSizer from "react-virtualized-auto-sizer"
import * as d3 from 'd3' 
import { sankeyCircular, sankeyJustify } from 'd3-sankey-circular'
import {GenerateRandomKey} from '../../util';


export default function Sankey(props) {
    const {getWorkflowSankeyMetrics, revision} = props

    const [links, setLinks] = useState([])
    const [nodes, setNodes] = useState([])
    const [load, setLoad] = useState(false)
    const [ini, setIni] = useState(true)

    useEffect(()=>{
        async function fetchMet() {
            let resp = await getWorkflowSankeyMetrics(revision)
            return resp.states
        }
        async function gatherMetrics(){
            setLoad(true)

            let n = []
            let l = []
    
            let states = await fetchMet()
            // Fill the nodes before doing the links so we can search up the states
            for(var i=0; i < states.length; i++) {
                n.push({name: states[i].name})
            }
            // loop success and failures to end state
            n.push({name: "end"})

            var failure = 0
            var success = 0
            // Write the links
            for(i=0; i < states.length; i++) {
                let outcomes = states[i].outcomes
                let source = states[i].name
                let invokers = states[i].invokers

                if(invokers.start){
                    let tpos = n.map((obj)=>{return obj.name}).indexOf("start")
                    if(tpos === -1) {
                        n.push({name:"start"}) 
                    }
                    l.push({source:"start", target: source, value: invokers.start})
                }

                if(outcomes.success !== 0){
                    let tpos = n.map((obj)=>{return obj.name}).indexOf("success")
                    if (tpos === -1) {
                        n.push({name: "success"})
                    }
                    l.push({source: source, target: "success", value: outcomes.success})
                    success += outcomes.success
                }
                if(outcomes.failure !== 0) {
                    let tpos = n.map((obj)=>{return obj.name}).indexOf("failure")
                    if (tpos === -1) {
                        n.push({name: "failure"})
                    }
                    l.push({source: source, target: "failure", value: outcomes.failure})
                    failure += outcomes.failure
                }

                if(outcomes.transitions) {
                    for(const state in outcomes.transitions){
                        l.push({source: source, target: state, value: outcomes.transitions[state]})
                    }
                }
            }

            if(success !== 0){
                l.push({source:"success", target: "end", value: success})
            }
            if(failure !== 0) {
                l.push({source:"failure", target: "end", value: failure})
            }
            if(states.length > 0) {
                setLinks(l)
                setNodes(n)
            }
        
            setLoad(false)
        }
        if(ini) {
            gatherMetrics()
            setIni(false)
        }
    },[getWorkflowSankeyMetrics, revision, ini])

    return(
        <div style={{height:"90%", width:"90%", minHeight:"300px", margin:"auto", marginTop:"20px", overflow:"hidden"}}>
            {
                load ? "":
                <AutoSizer>
                    {(dim)=> {
                        if(nodes.length > 0 && links.length > 0) {
                            return(
                                <SankeyDiagram nodes={nodes} links={links} height={dim.height-20} width={dim.width-40} />
                            )
                        }

                        return(
                            <div style={{textAlign:"center", paddingTop:"10px", fontSize:"11pt",  height:dim.height-20, width: dim.width}}>
                                No Metrics are found to draw the sankey.
                            </div>
                        )
                    }}
                </AutoSizer>
            }
        </div>
    )
}

function SankeyDiagram(props) {

    const {height, width, nodes, links} = props
    const margin = { top: 30, right: 30, bottom: 30, left: 30}

    useEffect(()=>{
        document.getElementById("sankey-graph").innerHTML = ""
        var sankey = sankeyCircular()
                        .nodeWidth(10)
                        .nodePaddingRatio(0.7)
                        .size([width, height])
                        .nodeId(function (d) {
                            return d.name;
                        })
                        .iterations(32)
                        .circularLinkGap(2)
                        .nodeAlign(sankeyJustify)
                        
        var svg = d3.select("#sankey-graph").append("svg")
                    .attr("width", width + margin.left + margin.right)
                    .attr("height", height + margin.top + margin.bottom);


        var defs = svg.append("defs")

        // var lg = defs.append("linearGradient")
        // .attr("id", "gradient")
        // .attr("x1", "0%")
        // .attr("y1", "0%")

        // var stop1 = lg.append("stop")
        // .attr("offset", "0%")
        // .style("stop-color", "#00bc9b")
        // .style("stop-opacity", "0.5")

        // var stop2 = lg.append("stop")
        // .attr("offset", "100%")
        // .style("stop-color", "#5eaefd")
        // .style("stop-opacity", "0.5")


        var g = svg.append("g")
                    .attr("transform", "translate(" + margin.left + "," + margin.top + ")")
        var linkG = g.append("g")
                    .attr("class", "links")
                    .attr("fill", "none")
                    .attr("stroke-opacity", 0.25)
                    .selectAll("path");
        var nodeG = g.append("g")
                    .attr("class", "nodes")
                    .attr("font-family", "sans-serif")
                    .attr("font-size", 10)
                    .selectAll("g");

        let sankeyData = sankey({nodes: nodes, links: links});
        let sankeyNodes = sankeyData.nodes;
        let sankeyLinks = sankeyData.links;

        // let depthExtent = d3.extent(sankeyNodes, function (d) { return d.depth; });

        var nodeColour = d3.scaleSequential(d3.interpolateCool)
        .domain([0,width]);
    
        var node = nodeG.data(sankeyNodes)
          .enter()
          .append("g");
    
        node.append("rect")
          .attr("x", function (d) { return d.x0; })
          .attr("y", function (d) { return d.y0; })
          .attr("height", function (d) { return d.y1 - d.y0; })
          .attr("width", function (d) { return d.x1 - d.x0; })
          .style("fill", function (d) {
              return nodeColour((d.x0 + d.x1 + d.y0 + d.y1)%width); 
          })
          .style("opacity", 0.85)
    
        node.append("text")
          .attr("x", function (d) { return (d.x0 + d.x1) / 2; })
          .attr("y", function (d) { return d.y0 - 12; })
          .attr("dy", "0.35em")
          .attr("text-anchor", "middle")
          .text(function (d) { return d.name; });
    
        node.append("title")
          .text(function (d) { return d.name + "\n" + (d.value); });
    
        var link = linkG.data(sankeyLinks)
          .enter()
          .append("g")

        link.append("path") 
        .attr("d", function(linkz){
            console.log(linkz
                );
            return linkz.path;
        })
        .attr("class", "sankey-link")
        .style("stroke-width", function (d) { return Math.max(1, d.width); })
        .style("stroke", function(linkz) {
            
            let id = GenerateRandomKey()
            
            let lingrad = defs.append("linearGradient")
            .attr("id", id)
            .attr("x1", linkz.source.x0)
            .attr("y1", linkz.source.y0)
            .attr("x2", linkz.target.x0)
            .attr("y2", linkz.target.y0)
            .attr("gradientUnits", "userSpaceOnUse")
            
            let s1 = lingrad.append("stop")
            .attr("offset", "0")
            
            let s2 = lingrad.append("stop")
            .attr("offset", "1")
            
            let sourceColour = (linkz.source.x0 + linkz.source.x1 + linkz.source.y0 + linkz.source.y1)%width
            let targetColour = (linkz.target.x0 + linkz.target.x1 + linkz.target.y0 + linkz.target.y1)%width

            // console.log(`Source: ${sourceColour}, Target: ${targetColour} ->`, linkz);

            s1.attr("stop-color", nodeColour(sourceColour))
            s2.attr("stop-color", nodeColour(targetColour))
            
            return `url(#${id})`
        })
 
        
        link.append("title")
        .text(function(d) { return d.source.name + " → " + d.target.name + "\n" + d.value; });
    

    },[height, width, nodes, links, margin.bottom, margin.left, margin.right, margin.top])

    return <div id="sankey-graph" style={{height: height, width:width}}/>
}

