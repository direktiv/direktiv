import {
  VariableNamespace,
  localVariableNamespace,
} from "../../../schema/primitives/variable";

import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { Preview } from "../SmartInput/Preview";
import { TreePicker } from "../TreePicker";
import { useCallback } from "react";
import { usePageEditorPanel } from "../../EditorPanelProvider";

export const VariableInput = ({
  value,
  onUpdate,
  blacklist,
  id,
  placeholder,
}: {
  onUpdate: (value: string) => void;
  value: string;
  id?: string;
  placeholder: string;
  blacklist?: VariableNamespace[];
}) => {
  const { panel } = usePageEditorPanel();

  const preventSubmit = useCallback((path: string[]) => {
    if (path[0] === localVariableNamespace && path.length > 1) return false;
    if (path.length > 2) return false;
    return true;
  }, []);

  if (!panel) return null;

  const { variables: allVariables } = panel;

  const variables = Object.fromEntries(
    Object.entries(allVariables).filter(
      ([key]) => !(blacklist as string[])?.includes(key)
    )
  );

  return (
    <InputWithButton>
      <Input
        id={id}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onUpdate(event.target.value)}
      />
      <TreePicker
        tree={variables}
        onSubmit={(variable) => onUpdate(variable)}
        preview={(path) => <Preview path={path} />}
        preventSubmit={preventSubmit}
      />
    </InputWithButton>
  );
};
