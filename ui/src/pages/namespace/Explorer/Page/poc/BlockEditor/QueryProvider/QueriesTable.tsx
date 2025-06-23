import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Dialog } from "~/design/Dialog";
import { DialogTrigger } from "@radix-ui/react-dialog";
import { ListContextMenu } from "~/components/ListContextMenu";
import { ModalWrapper } from "~/components/ModalWrapper";
import { Plus } from "lucide-react";
import { QueryForm } from "./QueryForm";
import { QueryType } from "../../schema/procedures/query";
import { useState } from "react";

type QueriesTableProps = {
  defaultValue: QueryType[];
  onChange: (newValues: QueryType[]) => void;
};

const formId = "query-form";

export const QueriesTable = ({ defaultValue, onChange }: QueriesTableProps) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [queries, setQueries] = useState(defaultValue);
  const [editIndex, setEditIndex] = useState<number>();

  const addQuery = (query: QueryType) => {
    const newQueries = [...queries, query];
    setQueries(newQueries);
    onChange(newQueries);
  };

  const updateQuery = (index: number, query: QueryType) => {
    const newQueries = queries.map((q, i) => (i === index ? query : q));
    setQueries(newQueries);
    onChange(newQueries);
  };

  const moveQuery = (srcIndex: number, targetIndex: number) => {
    const newQueries = [...queries];
    const [movedItem] = newQueries.splice(srcIndex, 1);
    if (!movedItem) throw new Error("Invalid source index");
    newQueries.splice(targetIndex, 0, movedItem);
    setQueries(newQueries);
    onChange(newQueries);
  };

  const deleteQuery = (index: number) => {
    const newQueries = queries.filter((_, i) => i !== index);
    setQueries(newQueries);
    onChange(newQueries);
  };

  const handleSubmit = (query: QueryType) => {
    setDialogOpen(false);
    if (editIndex === undefined) {
      addQuery(query);
    } else {
      updateQuery(editIndex, query);
    }
    setEditIndex(undefined);
  };

  const formValues = editIndex !== undefined ? queries[editIndex] : undefined;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <Card noShadow>
        <Table>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell colSpan={2}>
                {queries.length} Queries
              </TableHeaderCell>
              <TableHeaderCell className="w-60 text-right">
                <DialogTrigger asChild>
                  <Button icon variant="outline" size="sm">
                    <Plus />
                    add Query
                  </Button>
                </DialogTrigger>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {queries.map(({ id, url }, index, srcArray) => {
              const canMoveDown = index < srcArray.length - 1;
              const canMoveUp = index > 0;
              const onMoveUp = canMoveUp
                ? () => {
                    moveQuery(index, index - 1);
                  }
                : undefined;
              const onMoveDown = canMoveDown
                ? () => {
                    moveQuery(index, index + 1);
                  }
                : undefined;

              const onDelete = () => {
                deleteQuery(index);
              };

              return (
                <TableRow
                  key={id}
                  className="cursor-pointer"
                  onClick={() => {
                    setDialogOpen(true);
                    setEditIndex(index);
                  }}
                >
                  <TableCell>{id}</TableCell>
                  <TableCell>{url}</TableCell>
                  <TableCell className="text-right">
                    <ListContextMenu
                      onDelete={onDelete}
                      onMoveDown={onMoveDown}
                      onMoveUp={onMoveUp}
                    />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Card>
      <ModalWrapper
        formId={formId}
        title="title"
        onCancel={() => {
          setDialogOpen(false);
          setEditIndex(undefined);
        }}
      >
        <QueryForm
          formId={formId}
          onSubmit={handleSubmit}
          defaultValues={formValues}
        />
      </ModalWrapper>
    </Dialog>
  );
};
