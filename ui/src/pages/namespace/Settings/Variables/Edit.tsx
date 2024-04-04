import { VarFormUpdateSchemaType, VarSchemaType } from "~/api/variables/schema";

import { DialogContent } from "~/design/Dialog";
import { Generic } from "./Create";
import { useUpdateVar } from "~/api/variables/mutate/update";
import { useVarDetails } from "~/api/variables/query/details";

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
};

const Edit = ({ item, onSuccess }: EditProps) => {
  const { data, isSuccess } = useVarDetails(item.id);
  const { mutate: updateVar } = useUpdateVar({
    onSuccess,
  });

  const onMutate = (data: VarFormUpdateSchemaType) => {
    updateVar({
      id: item.id,
      ...data,
    });
  };

  return (
    <DialogContent>
      {isSuccess && (
        <Generic
          onMutate={onMutate}
          defaultValues={{
            name: data.data.name,
            data: data.data.data,
            mimeType: data.data.mimeType,
          }}
        />
      )}
    </DialogContent>
  );
};

export default Edit;
