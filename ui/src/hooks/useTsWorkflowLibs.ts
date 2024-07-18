import { useMemo } from "react";
import workflowTsDefinition from "~/assets/ts/workflow.d.ts?raw";

const useTsWorkflowLibs = (enabled: boolean) =>
  useMemo(
    () =>
      enabled
        ? [
            {
              content: workflowTsDefinition,
            },
          ]
        : [],
    [enabled]
  );

export default useTsWorkflowLibs;
