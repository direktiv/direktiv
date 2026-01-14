import { Edge, Node, Position, isNode } from "reactflow";
import { Orientation, State } from "./types";

import { Workflow } from "~/api/instances/schema";
import dagre from "dagre";

const defaultEdgeType = "default";

export const getLayoutedElements = (
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

export function generateElements(
  getLayoutedElements: (
    incomingEles: (Node | Edge)[],
    orientation: Orientation
  ) => (Node | Edge)[],
  value: Workflow,
  flow: string[],
  status: "pending" | "complete" | "failed",
  orientation: Orientation
) {
  const newElements: (Node | Edge)[] = [];
  if (!value) return [];

  const statesParent = (value as Workflow).states as unknown;
  let rawStates: Record<string, State> = {};

  if (statesParent && typeof statesParent === "object") {
    if ("state" in (statesParent as Record<string, unknown>)) {
      const s = (statesParent as { state: Record<string, State> }).state;
      if (s && typeof s === "object") rawStates = s;
    } else {
      rawStates = statesParent as Record<string, State>;
    }
  }

  const states = Object.values(rawStates) as State[];

  let isFirst = true;
  let lastNode: State | null = null;

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
    if (isFirst) {
      isFirst = false;
      const startId = value.start?.state ?? state.id;

      newElements.push({
        id: `startNode-${startId}`,
        source: "startNode",
        target: startId,
        type: defaultEdgeType,
        animated: state.visited,
      });
    }

    // create state node
    const stateNode: Node = {
      id: state.id,
      position,
      data: {
        label: state.id,
        type: state.type,
        state,
        functions: value.functions,
        wasExecuted: state.visited,
        orientation,
      },
      type: "state",
    };
    newElements.push(stateNode);

    // create edge to next state
    const sourceId = state.id;
    const outgoingTargets = new Set<string>();
    state.transitions.forEach((t) => t && outgoingTargets.add(t));
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
          state.visited &&
          states.find((state) => state.id === targetId)?.visited, // line gets animated if state before and after were visited
      });
    }

    // create end edge
    lastNode = states[states.length - 1] as State;
    const lastNodeId = lastNode?.id ?? "";

    if (state === lastNode) {
      newElements.push({
        id: `${lastNodeId}-endNode`,
        source: lastNodeId,
        target: "endNode",
        type: defaultEdgeType,
        animated: state.visited && status === "complete",
      });
    }
  }

  // create end node
  const reachedEnd = lastNode?.visited && status === "complete";

  newElements.push({
    id: "endNode",
    type: "end",
    data: { label: "", wasExecuted: reachedEnd, orientation },
    position,
  });

  return getLayoutedElements(newElements, orientation);
}
