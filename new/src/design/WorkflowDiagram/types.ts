export type State = {
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

type NonEmptyArray<T> = [T, ...T[]];

export type Workflow = {
  states: NonEmptyArray<State>;
  start: {
    state: string;
  };
  functions: object[];
};

export type Orientation = "horizontal" | "vertical";
