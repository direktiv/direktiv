export type State = {
  name: string;
  start: boolean;
  finish: boolean;
  visited: boolean;
  failed: boolean;
  transitions: string[];
  id: string;
  type: string;
  events: {
    transition: string;
  }[];
  conditions: {
    transition: string;
  }[];
  catch: {
    x: string;
    y: string;
    transition: string;
  }[];
  transition: string;
  defaultTransition: string;
};

export type Orientation = "horizontal" | "vertical";
