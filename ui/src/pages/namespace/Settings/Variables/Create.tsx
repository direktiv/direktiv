import { DialogContent } from "~/design/Dialog";
import { VariableForm } from "./Form";
import { useCreateVar } from "~/api/variables/mutate/create";

const defaultMimeType = "application/json";

type CreateProps = { onSuccess: () => void };

const Create = ({ onSuccess }: CreateProps) => {
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  return (
    <DialogContent>
      <VariableForm
        onMutate={createVar}
        defaultValues={{
          name: "",
          data: "",
          mimeType: defaultMimeType,
        }}
      />
    </DialogContent>
  );
};

export default Create;
