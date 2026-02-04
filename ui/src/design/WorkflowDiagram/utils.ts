import { Edge, Node, Position, isNode } from "reactflow";

import { InstanceFlowResponse } from "~/api/instances/schema";
import { Orientation } from "./types";
import dagre from "dagre";

const defaultEdgeType = "default";

const createLayoutedElements = (
  incomingEles: (Edge | Node)[],
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
  value: InstanceFlowResponse,
  status: "pending" | "complete" | "failed",
  orientation: Orientation
) {
  const newElements: (Node | Edge)[] = [];
  if (!value) return [];

  const states = Object.values(value.data);

  if (states.length === 0) return [];

  // create start node
  newElements.push({
    id: "startNode",
    position,
    data: { label: "", wasExecuted: status !== "pending", orientation },
    type: "start",
    sourcePosition: Position.Right,
  });

  // loop through all the state nodes
  for (const state of states) {
    // create start edge
    if (state.start === true) {
      newElements.push({
        id: `startNode-${state.name}`,
        source: "startNode",
        target: state.name,
        type: defaultEdgeType,
        animated: state.visited,
      });
    }

    // create state node
    const stateNode: Node = {
      id: state.name,
      position,
      data: {
        type: "function",
        label: state.name,
        state,
        wasExecuted: state.visited,
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
          state.visited && states.find((s) => s.name === targetId)?.visited,
      });
    }

    // create end edges
    if (state.finish === true) {
      newElements.push({
        id: `${state.name}-endNode`,
        source: state.name,
        target: "endNode",
        type: defaultEdgeType,
        animated: state.visited && status === "complete",
      });
    }
  }

  newElements.push({
    id: "endNode",
    type: "end",
    data: { label: "", wasExecuted: status === "complete", orientation },
    position,
  });

  return createLayoutedElements(newElements, orientation);
}
