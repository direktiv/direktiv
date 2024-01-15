import { Eye, EyeOff } from "lucide-react";
import { FC, useState } from "react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";

type SecretInputProps = {
  secret: string;
};

const SecretInput: FC<SecretInputProps> = ({ secret }) => {
  const [revealSecret, setRevealSecret] = useState(false);
  return (
    <InputWithButton>
      <Input
        value={secret}
        readOnly
        type={revealSecret ? "text" : "password"}
      />
      <Button
        variant="ghost"
        onClick={() => setRevealSecret(!revealSecret)}
        icon
      >
        {revealSecret ? <EyeOff /> : <Eye />}
      </Button>
    </InputWithButton>
  );
};

export default SecretInput;
