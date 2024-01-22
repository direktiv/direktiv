import { FC, useState } from "react";

import Button from "~/design/Button";
import { EnvironementVariableSchemaType } from "~/api/services/schema/services";
import Input from "~/design/Input";
import { Plus } from "lucide-react";

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
}) => {
  // const [name, setName] = useState<string>(item?.name || "");
  // const [value, setValue] = useState<string>(item?.value || "");
  const [state, setState] = useState<EnvironementVariableSchemaType>(
    item || {
      name: "",
      value: "",
    }
  );

  const handleChange = (object: Partial<EnvironementVariableSchemaType>) => {
    const newState = { ...state, ...object };
    setState(newState);
    onUpdate && onUpdate(newState);
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
        placeholder="VALUE"
      ></Input>
      {onAdd && (
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            onAdd && onAdd(state);
          }}
        >
          <Plus />
        </Button>
      )}
    </div>
  );
};
