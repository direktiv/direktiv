import { Eye, EyeOff } from "lucide-react";
import { FC, useState } from "react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";

type PasswordInputProps = {
  password: string;
};

const PasswordInput: FC<PasswordInputProps> = ({ password }) => {
  const [revealPassword, setRevealPassword] = useState(false);
  return (
    <InputWithButton>
      <Input
        value={password}
        readOnly
        type={revealPassword ? "text" : "password"}
      />
      <Button
        variant="ghost"
        onClick={() => setRevealPassword(!revealPassword)}
        icon
      >
        {revealPassword ? <EyeOff /> : <Eye />}
      </Button>
    </InputWithButton>
  );
};

export default PasswordInput;
