import { End, Start, State } from "./nodes";
import ReactFlow, { Edge, MiniMap, Node, useReactFlow } from "reactflow";
import { useEffect, useMemo } from "react";

interface ZoomPanDiagramProps {
  elements: (Edge | Node)[];
  disabled: boolean;
}

const nodeTypes = {
  state: State,
  start: Start,
  end: End,
};

export function ZoomPanDiagram(props: ZoomPanDiagramProps) {
  const { elements, disabled } = props;
  const { fitView } = useReactFlow();

  const sep: [Node[], Edge[]] = useMemo(() => {
    const nodes: Node[] = elements.filter(
      (ele: Node | Edge) => !!(ele as Node).position
    ) as Node[];

    const edges: Edge[] = elements.filter(
      (ele: Node | Edge) => !!(ele as Edge).source
    ) as Edge[];
    return [nodes, edges];
  }, [elements]);
  useEffect(() => {
    fitView();
  }, [fitView]);
  return (
    <ReactFlow
      edges={sep[1]}
      nodes={sep[0]}
      nodeTypes={nodeTypes}
      nodesDraggable={disabled}
      nodesConnectable={disabled}
      elementsSelectable={disabled}
    >
      <MiniMap nodeColor={() => "#4497f5"} />
    </ReactFlow>
  );
}
