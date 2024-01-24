import { FC, KeyboardEvent, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { EnvironementVariableSchemaType } from "~/api/services/schema/services";
import Input from "~/design/Input";

type EnvItemFormProps = {
  item?: EnvironementVariableSchemaType;
  onAdd?: (value: EnvironementVariableSchemaType) => void;
  onUpdate?: (value: EnvironementVariableSchemaType) => void;
  onDelete?: () => void;
};

export const EnvItemForm: FC<EnvItemFormProps> = ({
  item,
  onAdd,
  onUpdate,
  onDelete,
}) => {
  const emptyItem = {
    name: "",
    value: "",
  };

  const [state, setState] = useState<EnvironementVariableSchemaType>(
    item || emptyItem
  );

  const handleChange = (object: Partial<EnvironementVariableSchemaType>) => {
    const newState = { ...state, ...object };
    setState(newState);
    onUpdate && onUpdate(newState);
  };

  const isValid = state?.name && state?.value;

  const handleAdd = () => {
    onAdd && onAdd(state);
    setState(emptyItem);
  };

  const handleDelete = () => {
    onDelete && onDelete();
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    if (!isValid) return;
    if (event.key === "Enter") {
      event.preventDefault();
      handleAdd();
    }
  };

  return (
    <div className="flex gap-3">
      <Input
        value={state.name}
        onChange={(event) => handleChange({ name: event.target.value })}
        placeholder="NAME"
      ></Input>
      <Input
        value={state.value}
        onChange={(event) => handleChange({ value: event.target.value })}
        onKeyDown={handleKeyDown}
        placeholder="VALUE"
      ></Input>
      {onAdd && (
        <Button
          type="button"
          variant="outline"
          onClick={() => handleAdd()}
          disabled={!isValid}
        >
          <Plus />
        </Button>
      )}
      {onDelete && (
        <Button type="button" variant="outline" onClick={() => handleDelete()}>
          <X />
        </Button>
      )}
    </div>
  );
};
