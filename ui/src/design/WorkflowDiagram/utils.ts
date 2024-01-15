import { ConnectionLineType, Edge, Node, Position, isNode } from "reactflow";
import { Orientation, Workflow } from "./types";

import dagre from "dagre";

const defaultEdgeType = ConnectionLineType.Bezier;

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
      if (el.id === "startNode" || el.id === "endNode") {
        dagreGraph.setNode(el.id, { width: 40, height: 40 });
      } else {
        dagreGraph.setNode(el.id, { width: 100, height: 40 });
      }
    } else {
      if (el.source === "startNode") {
        dagreGraph.setEdge(el.source, el.target, { width: 0, height: 20 });
      } else if (el.source === "endNode") {
        dagreGraph.setEdge(el.source, el.target, { width: 30, height: 20 });
      } else {
        dagreGraph.setEdge(el.source, el.target, { width: 60, height: 60 });
      }
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
  value: Workflow | null,
  flow: string[],
  status: "pending" | "complete" | "failed",
  orientation: Orientation
) {
  const newElements: (Node | Edge)[] = [];

  if (!value) return [];

  if (value.states) {
    for (let i = 0; i < value.states.length; i++) {
      let transitions = false;
      // check if starting element
      if (i === 0) {
        // starting element so create an edge to the state
        if (value.start && value.start.state) {
          newElements.push({
            id: `startNode-${value.start.state}`,
            source: "startNode",
            target: value.start.state,
            type: defaultEdgeType,
          });
        } else {
          newElements.push({
            id: `startNode-${value.states[i]?.id}`,
            source: "startNode",
            target: value.states[i].id,
            type: defaultEdgeType,
          });
        }
      }

      // push new state
      newElements.push({
        id: value.states[i]?.id || "",
        position,
        data: {
          label: value.states[i]?.id || "",
          type: value.states[i]?.type || "",
          state: value.states[i],
          functions: value.functions,
          wasExecuted: flow.includes(value.states[i]?.id || ""),
          orientation,
        },
        type: "state",
      });

      // check if the state has events
      if (value.states[i]?.events) {
        for (let j = 0; j < (value.states[i]?.events.length || 0); j++) {
          if (value.states[i]?.events[j]?.transition) {
            transitions = true;
            newElements.push({
              id: `${value.states[i]?.id}-${value.states[i]?.events[j]?.transition}`,
              source: value.states[i]?.id || "",
              target: value.states[i]?.events[j]?.transition || "",
              animated: false,
              type: defaultEdgeType,
            });
          }
        }
      }

      // Check if the state has conditions
      if (value.states[i]?.conditions) {
        for (let y = 0; y < (value.states[i]?.conditions?.length || 0); y++) {
          if (value.states[i]?.conditions[y]?.transition) {
            newElements.push({
              id: `${value.states[i]?.id}-${value.states[i]?.conditions[y]?.transition}-${i}-${y}`,
              source: value.states[i]?.id || "",
              target: value.states[i]?.conditions[y]?.transition || "",
              animated: false,
              type: defaultEdgeType,
            });
            transitions = true;
          }
        }
      }

      // Check if state is catching things to transition to
      if (value.states[i]?.catch) {
        for (let x = 0; x < (value.states[i]?.catch?.length || 0); x++) {
          if (value.states[i]?.catch[x]?.transition) {
            transitions = true;

            newElements.push({
              id: `${value.states[i]?.id}-${value.states[i]?.catch[x]?.transition}`,
              source: value.states[i]?.id || "",
              target: value.states[i]?.catch[x]?.transition || "",
              animated: false,
              type: defaultEdgeType,
            });
          }
        }
      }

      // check if transition and create edge to hit new state
      if (value.states[i]?.transition) {
        transitions = true;

        newElements.push({
          id: `${value.states[i]?.id}-${value.states[i]?.transition}`,
          source: value.states[i]?.id || "",
          target: value.states[i]?.transition || "",
          animated: false,
          type: defaultEdgeType,
        });
      } else if (value.states[i]?.defaultTransition) {
        transitions = true;

        newElements.push({
          id: `${value.states[i]?.id}-${value.states[i]?.defaultTransition}`,
          source: value.states[i]?.id || "",
          target: value.states[i]?.defaultTransition || "",
          animated: false,
          type: defaultEdgeType,
        });
      } else {
        transitions = true;
        newElements.push({
          id: `${value.states[i]?.id}-endNode`,
          source: value.states[i]?.id || "",
          target: `endNode`,
          animated: false,
          type: defaultEdgeType,
        });
      }

      if (!transitions) {
        // no transition add end state
        newElements.push({
          id: `${value.states[i]?.id}-endNode`,
          source: value.states[i]?.id || "",
          target: `endNode`,
          type: defaultEdgeType,
        });
      }
    }

    const hasStarted = status === "failed" || status === "complete";

    // push start node
    newElements.push({
      id: "startNode",
      position,
      data: { label: "", wasExecuted: hasStarted, orientation },
      type: "start",
      sourcePosition: Position.Right,
    });

    // Check flow array change edges to green if it passed
    if (flow) {
      // check flow for transitions
      for (let i = 0; i < flow.length; i++) {
        let noTransition = false;
        for (let j = 0; j < newElements.length; j++) {
          // handle start node
          const item = newElements[j] && (newElements[j] as Edge);
          if (item && item.source === "startNode" && item.target === flow[i]) {
            // connection between start and first state
            (newElements[j] as Edge).animated = true;
          }

          if (item && item.target === flow[i] && item.source === flow[i - 1]) {
            // connection between two states
            (newElements[j] as Edge).animated = true;
          } else if (item && item.id === flow[i]) {
            if (
              !item.data.state.transition ||
              !item.data.state.defaultTransition
            ) {
              noTransition = true;

              if (item.data.state.catch) {
                for (
                  let y = 0;
                  y < (newElements[j] as Edge).data.state.catch.length;
                  y++
                ) {
                  if ((newElements[j] as Edge).data.state.catch[y].transition) {
                    noTransition = false;
                    if (
                      (newElements[j] as Edge).data.label ===
                      flow[flow.length - 1]
                    ) {
                      noTransition = true;
                    }
                  }
                }
              }
            }
          }
        }

        if (noTransition) {
          // transition to end state
          // check if theres more flow if not its the end node
          if (!flow[i + 1]) {
            for (let j = 0; j < newElements.length; j++) {
              if (
                (newElements[j] as Edge).source === flow[i] &&
                (newElements[j] as Edge).target === "endNode" &&
                (status === "complete" || status === "failed")
              ) {
                // connection between the last state and end node
                (newElements[j] as Edge).animated = true;
              }
            }
          }
        }
      }
    }

    const reachedEnd = newElements.some(
      // newElements is typed as an array of Node | Edge but this is not
      // quite true. The attribute target is added to Edge as a helper
      // when refactoring this should be fixed
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      (x) => x.target === "endNode" && x.animated === true
    );

    // push end node
    newElements.push({
      id: "endNode",
      type: "end",
      data: { label: "", wasExecuted: reachedEnd, orientation },
      position,
    });
  }
  return getLayoutedElements(newElements, orientation);
}
