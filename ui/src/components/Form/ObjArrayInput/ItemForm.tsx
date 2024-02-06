import { FC, KeyboardEvent, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { UnknownObjectOfStrings } from "./types";

type ItemFormProps = {
  item?: UnknownObjectOfStrings;
  onAdd?: (value: UnknownObjectOfStrings) => void;
  onUpdate?: (value: UnknownObjectOfStrings) => void;
  onDelete?: () => void;
};

export const ItemForm: FC<ItemFormProps> = ({
  item,
  onAdd,
  onUpdate,
  onDelete,
}) => {
  const emptyItem = {
    name: "",
    value: "",
  };

  const [state, setState] = useState<UnknownObjectOfStrings>(item || emptyItem);

  const handleChange = (object: Partial<UnknownObjectOfStrings>) => {
    const newState = { ...state, ...object };
    setState(newState);
    onUpdate && onUpdate(newState);
  };

  const isValid = state?.name && state?.value;

  const handleAdd = () => {
    if (!onAdd) return;
    onAdd(state);
    setState(emptyItem);
  };

  const handleDelete = () => {
    onDelete && onDelete();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!isValid) return;
    if (event.key === "Enter") {
      if (!onAdd) return;
      event.preventDefault();
      event.currentTarget.blur();
      handleAdd();
    }
  };

  return (
    <div className="flex gap-3" data-testid="env-item-form">
      {Object.keys(state).map((key) => (
        <Input
          key={key}
          value={state[key]}
          onChange={(event) => handleChange({ [key]: event.target.value })}
          onKeyDown={handleKeyDown}
          placeholder={key}
          data-testid={`env-${key}`}
        />
      ))}
      {/* <Input
        value={state.name}
        onChange={(event) => handleChange({ name: event.target.value })}
        onKeyDown={handleKeyDown}
        placeholder={t(
          "pages.explorer.service.editor.form.envs.namePlaceholder"
        )}
        data-testid="env-name"
      />
      <Input
        value={state.value}
        onChange={(event) => handleChange({ value: event.target.value })}
        onKeyDown={handleKeyDown}
        placeholder={t(
          "pages.explorer.service.editor.form.envs.valuePlaceholder"
        )}
        data-testid="env-value"
      /> */}
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
