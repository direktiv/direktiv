import { Edge, Node, Position, isNode } from "reactflow";
import { Orientation, State } from "./types";

import { DiagramElementStatus } from "./nodes";
import { InstanceFlowSchemaType } from "~/api/instances/schema";
import dagre from "dagre";

const defaultEdgeType = "default";

type DiagramNodeData = {
  type?: "function";
  label: string;
  status: DiagramElementStatus;
  orientation: Orientation;
};

const createLayoutedElements = (
  incomingEles: (Edge | Node<DiagramNodeData>)[],
  orientation: Orientation = "vertical"
) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  const isHorizontal = orientation === "horizontal";

  dagreGraph.setGraph({ rankdir: isHorizontal ? "lr" : "tb" });

  incomingEles.forEach((el) => {
    if (isNode(el)) {
      dagreGraph.setNode(el.id, {
        width: el.id === "startNode" || el.id === "endNode" ? 40 : 100,
        height: el.id === "startNode" || el.id === "endNode" ? 40 : 40,
      });
    } else {
      dagreGraph.setEdge(el.source, el.target, { width: 60, height: 60 });
    }
  });

  dagre.layout(dagreGraph);

  return incomingEles.map((el) => {
    if (isNode(el)) {
      const nodeWithPosition = dagreGraph.node(el.id);
      el.targetPosition = isHorizontal ? Position.Left : Position.Top;
      el.sourcePosition = isHorizontal ? Position.Right : Position.Bottom;
      el.position = {
        x: nodeWithPosition.x,
        y: nodeWithPosition.y,
      };
    }
    return el;
  });
};

const position = { x: 0, y: 0 };

export function createElements(
  value: InstanceFlowSchemaType,
  instanceStatus: "pending" | "complete" | "failed",
  orientation: Orientation
) {
  const newElements: (Node<DiagramNodeData> | Edge)[] = [];
  if (!value) return [];

  const visitedStates = value.flow || [];

  const states = value.states;

  if (states.length === 0) return [];

  const hasInstanceStarted = visitedStates.length > 0;

  // create start node
  newElements.push({
    id: "startNode",
    position,
    data: {
      label: "",
      status: hasInstanceStarted ? "complete" : "pending",
      orientation,
    },
    type: "start",
    sourcePosition: Position.Right,
  });

  const startState = states.find((s) => s && (s as State).start === true);
  const startId = startState ? startState.name : "";

  // loop through all the state nodes
  for (const [index, state] of states.entries()) {
    // create start edge
    if (index === 0) {
      newElements.push({
        id: `startNode-${startId}`,
        source: "startNode",
        target: startId,
        type: defaultEdgeType,
        animated: state.visited,
      });
    }

    const stateNode: Node<DiagramNodeData> = {
      id: state.name,
      position,
      data: {
        type: "function",
        label: state.name,
        status:
          (state.failed && "failed") ||
          (state.visited && "complete") ||
          "pending",
        orientation,
      },
      type: "state",
    };
    newElements.push(stateNode);

    // create edge to next state
    const sourceId = state.name;
    const outgoingTargets = new Set<string>();
    state.transitions?.forEach((t) => t && outgoingTargets.add(t));
    state.events?.forEach(
      (ev) => ev.transition && outgoingTargets.add(ev.transition)
    );
    state.conditions?.forEach(
      (cond) => cond.transition && outgoingTargets.add(cond.transition)
    );
    state.catch?.forEach(
      (c) => c.transition && outgoingTargets.add(c.transition)
    );
    if (state.transition) outgoingTargets.add(state.transition);
    else if (state.defaultTransition)
      outgoingTargets.add(state.defaultTransition);

    for (const targetId of outgoingTargets) {
      newElements.push({
        id: `${sourceId}-${targetId}`,
        source: sourceId,
        target: targetId,
        type: defaultEdgeType,
        animated:
          targetId ===
          visitedStates[visitedStates.findIndex((s) => s === sourceId) + 1],
      });
    }

    // create end edges
    if (state.finish === true) {
      newElements.push({
        id: `${state.name}-endNode`,
        source: state.name,
        target: "endNode",
        type: defaultEdgeType,
        animated:
          visitedStates[visitedStates.length - 1] === state.name &&
          instanceStatus === "complete",
      });
    }
  }

  newElements.push({
    id: "endNode",
    type: "end",
    data: {
      label: "",
      status: instanceStatus === "complete" ? "complete" : "pending",
      orientation,
    },
    position,
  });

  return createLayoutedElements(newElements, orientation);
}
