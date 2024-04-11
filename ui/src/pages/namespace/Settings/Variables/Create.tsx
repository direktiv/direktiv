import { CreateVariableForm } from "~/components/VariableForm/CreateForm";
import { useCreateVar } from "~/api/variables/mutate/create";

type CreateProps = { onSuccess: () => void; unallowedNames: string[] };

const Create = ({ onSuccess, unallowedNames }: CreateProps) => {
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  return (
    <CreateVariableForm onMutate={createVar} unallowedNames={unallowedNames} />
  );
};

export default Create;
