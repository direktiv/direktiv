import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { NewEventSchema, NewEventSchemaType } from "~/api/eventsv2/schema";
import { PlusCircle, Radio } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { exampleEvent } from "./exampleEvent";
import { useSendEvent } from "~/api/eventsv2/mutate/sendEvent";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const SendEvent = () => {
  const [body, setBody] = useState<string | undefined>(exampleEvent);
  const [isOpen, setIsOpen] = useState(false);
  const theme = useTheme();
  const { t } = useTranslation();

  const { mutate } = useSendEvent({
    onSuccess: () => setIsOpen(false),
  });

  const onSubmit: SubmitHandler<NewEventSchemaType> = (data) => {
    mutate(data);
  };

  const { handleSubmit } = useForm<NewEventSchemaType>({
    resolver: zodResolver(NewEventSchema),
    values: {
      body: body || "",
    },
  });

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button className="whitespace-nowrap" icon variant="primary">
          <PlusCircle />
          {t("pages.events.history.send.dialogTrigger")}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-xl md:max-w-2xl lg:max-w-3xl">
        <form
          data-testid="send-event-form"
          id="send-event"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
        >
          <DialogHeader>
            <DialogTitle>
              <Radio />
              {t("pages.events.history.send.dialogHeader")}
            </DialogTitle>
          </DialogHeader>
          <Card
            className="grow p-4 pl-0"
            background="weight-1"
            data-testid="event-create-card"
          >
            <div className="h-[500px]">
              <Editor
                value={body}
                language="json"
                theme={theme ?? undefined}
                onChange={(newData) => {
                  setBody(newData);
                }}
              />
            </div>
          </Card>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" variant="primary">
              {t("pages.events.history.send.submit")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default SendEvent;
