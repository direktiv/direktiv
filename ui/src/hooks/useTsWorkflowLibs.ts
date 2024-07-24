import * as monaco from "monaco-editor";

import { useMemo } from "react";
import workflowTsDefinition from "~/assets/ts/workflow.d.ts?raw";

type SetExtraLibsArgument = Parameters<
  typeof monaco.languages.typescript.typescriptDefaults.setExtraLibs
>[0];

const useTsWorkflowLibs = (enabled: boolean) =>
  useMemo<SetExtraLibsArgument>(
    () => (enabled ? [{ content: workflowTsDefinition }] : []),
    [enabled]
  );

export default useTsWorkflowLibs;
