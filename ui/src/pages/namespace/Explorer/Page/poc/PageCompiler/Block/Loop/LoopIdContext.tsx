import { FC, PropsWithChildren, createContext, useContext } from "react";

/**
 * This context provides the current index of a loop to its children.
 * It stores an object where each key is a loop ID and its value is
 * the current index. While each loop component only needs to provide
 * its own index, nested loops must also pass down their parent loops'
 * indices since React context will always return the value from the
 * closest parent provider.
 */
export type State = Record<string, number>;

const LoopIdContext = createContext<State | null>(null);

type LoopIdContextProviderProps = PropsWithChildren<{ value: State }>;

const LoopIdContextProvider: FC<LoopIdContextProviderProps> = ({
  children,
  value,
}) => <LoopIdContext.Provider value={value}>{children}</LoopIdContext.Provider>;

const useLoopIdContext = () => {
  const context = useContext(LoopIdContext);
  return context;
};

const useLoopIndex = () => useLoopIdContext() ?? undefined;

export { LoopIdContextProvider, useLoopIndex };
