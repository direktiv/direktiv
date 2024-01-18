import { FC, useState } from "react";

import Button from "~/design/Button";
import { EnvironementVariableSchemaType } from "~/api/services/schema/services";
import Input from "~/design/Input";
import { Plus } from "lucide-react";

type EnvItemFormProps = {
  item?: EnvironementVariableSchemaType;
  onAdd?: (value: EnvironementVariableSchemaType) => void;
};

export const EnvItemForm: FC<EnvItemFormProps> = ({ item, onAdd }) => {
  const [name, setName] = useState<string>(item?.name || "");
  const [value, setValue] = useState<string>(item?.value || "");

  return (
    <div className="flex gap-3">
      <Input
        value={name}
        onChange={(event) => setName(event.target.value)}
        placeholder="NAME"
      ></Input>
      <Input
        value={value}
        onChange={(event) => setValue(event.target.value)}
        placeholder="VALUE"
      ></Input>
      <Button
        type="button"
        variant="outline"
        onClick={() => {
          onAdd && onAdd({ name, value });
        }}
      >
        <Plus />
      </Button>
    </div>
  );
};
