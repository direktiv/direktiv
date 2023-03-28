import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../../design/Dialog";
import { PlusCircle, Trash } from "lucide-react";

import Button from "../../../design/Button";
import { NodeSchemaType } from "../../../api/tree/schema";
import { useDeleteNode } from "../../../api/tree/mutate/deleteNode";

const Delete = ({
  node,
  close,
}: {
  node: NodeSchemaType;
  close: () => void;
}) => {
  const { mutate, isLoading } = useDeleteNode({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> Delete
        </DialogTitle>
      </DialogHeader>
      <div>
        Are you sure you want to delete <b>{node.name}</b>? This can not be
        undone.
        {node.type === "directory" && (
          <div>All content of this directoy will be deleted as well.</div>
        )}
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button
          onClick={() => {
            mutate({ node });
          }}
          variant="destructive"
          loading={isLoading}
        >
          {!isLoading && <Trash />}
          Delete
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
