import { DefaultEdgeOptions, Edge, Node, Position, isNode } from "reactflow";
import { Orientation, State } from "./types";

import { Workflow } from "~/api/instances/schema";
import dagre from "dagre";

const defaultEdgeType = "default";

const getStates = (workflow: Workflow) => {
  const statesParent = (workflow as Workflow).states as unknown;
  let rawStates: Record<string, State> = {};

  if (statesParent && typeof statesParent === "object") {
    if ("state" in (statesParent as Record<string, unknown>)) {
      const s = (statesParent as { state: Record<string, State> }).state;
      if (s && typeof s === "object") {
        rawStates = s;
      }
    } else {
      rawStates = statesParent as Record<string, State>;
    }
  }

  return Object.values(rawStates) as State[];
};

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

const createEdges = (
  state: State,
  visitedStates: State[]
): Edge<DefaultEdgeOptions>[] => {
  const targetIds = new Set<string>();

  state.transitions.forEach((t) => t && targetIds.add(t));
  state.events?.forEach((ev) => ev.transition && targetIds.add(ev.transition));
  state.conditions?.forEach(
    (cond) => cond.transition && targetIds.add(cond.transition)
  );
  state.catch?.forEach((c) => c.transition && targetIds.add(c.transition));

  if (state.transition) {
    targetIds.add(state.transition);
  } else if (state.defaultTransition) {
    targetIds.add(state.defaultTransition);
  }

  const edges: Edge<DefaultEdgeOptions>[] = [];

  for (const targetId of targetIds) {
    edges.push({
      id: `${state.id}-${targetId}`,
      source: state.id,
      target: targetId,
      type: defaultEdgeType,
      animated:
        // line is animated if this state and the target were visited
        state.visited && visitedStates.some((state) => state.id === targetId),
    });
  }

  return edges;
};

export const createElements = (
  value: Workflow,
  status: "pending" | "complete" | "failed",
  orientation: Orientation
) => {
  const newElements: (Node | Edge)[] = [];
  if (!value) return [];

  const states = getStates(value);

  // create start node
  newElements.push({
    id: "startNode",
    position: { x: 0, y: 0 },
    data: { label: "", wasExecuted: status !== "pending", orientation },
    type: "start",
    sourcePosition: Position.Right,
  });

  // loop through all the state nodes
  for (const [index, state] of states.entries()) {
    // create start edge
    if (index === 0) {
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
      position: { x: 0, y: 0 },
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

    // create edges to next states
    const edges = createEdges(
      state,
      states.filter((state) => state.visited)
    );

    for (const edge of edges) {
      newElements.push(edge);
    }

    // create end edge
    if (index === states.length - 1) {
      newElements.push({
        id: `${state.id}-endNode`,
        source: state.id,
        target: "endNode",
        type: defaultEdgeType,
        animated: state.visited && status === "complete",
      });
    }
  }

  // create end node
  const lastNode = states[states.length - 1];
  const lastNodeWasExecuted = lastNode?.visited && status === "complete";

  newElements.push({
    id: "endNode",
    type: "end",
    data: { label: "", wasExecuted: lastNodeWasExecuted, orientation },
    position: { x: 0, y: 0 },
  });

  return createLayoutedElements(newElements, orientation);
};
