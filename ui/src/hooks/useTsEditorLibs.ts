import { useMemo } from "react";
import workflowTsDefinition from "~/assets/ts/workflow.d.ts?raw";

const useTsEditorLibs = (enabled: boolean) =>
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

export default useTsEditorLibs;
