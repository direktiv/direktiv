import { End, Start, State } from "./nodes";
import ReactFlow, {
  Background,
  Edge,
  MiniMap,
  Node,
  useNodesInitialized,
  useReactFlow,
} from "reactflow";
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
  const nodesInitialized = useNodesInitialized();

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
    if (nodesInitialized) {
      fitView();
    }
  }, [fitView, nodesInitialized]);

  return (
    <ReactFlow
      edges={sep[1]}
      nodes={sep[0]}
      nodeTypes={nodeTypes}
      nodesDraggable={!disabled}
      nodesConnectable={!disabled}
      elementsSelectable={!disabled}
      fitView={true}
    >
      <MiniMap />
      <Background />
    </ReactFlow>
  );
}
