import {
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";

import { VariableNamespace } from "../../../schema/primitives/variable";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;

export type Variables = {
  [keys in VariableNamespace]: Record<VariableId, DefinedValue>;
};

const defaultVariables: Variables = {
  loop: {},
  form: {},
  query: {},
};

type State = {
  variables: Variables;
  setVariables: Dispatch<SetStateAction<Variables>>;
};

const VariableContext = createContext<State | null>(null);

type VariableContextProviderProps = PropsWithChildren<
  Omit<State, "setVariables">
>;

const VariableContextProvider = ({
  children,
  variables: initialValue,
}: VariableContextProviderProps) => {
  const [variables, setVariables] = useState<Variables>(initialValue);
  return (
    <VariableContext.Provider value={{ variables, setVariables }}>
      {children}
    </VariableContext.Provider>
  );
};

const useVariableContext = () => {
  const context = useContext(VariableContext);
  return context;
};

const useVariables = () => useVariableContext()?.variables ?? defaultVariables;

export { VariableContextProvider, useVariables };
