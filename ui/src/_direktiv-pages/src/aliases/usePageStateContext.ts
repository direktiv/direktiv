import { usePageStateContext as fullPageStateContext } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";

export const usePageStateContext = () => {
  const state = fullPageStateContext();

  return {
    ...state,
    mode: "view" as const,
  };
};
