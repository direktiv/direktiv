import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  EventSchemaType,
  NewEventSchema,
  NewEventSchemaType,
} from "~/api/events/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { Calendar } from "lucide-react";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { useReplayEvent } from "~/api/events/mutate/replayEvent";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const EventEditorForm = ({ event }: { event: EventSchemaType }) => {
  const [body, setBody] = useState<string | undefined>(
    atob(event.cloudevent) || undefined
  );

  const { t } = useTranslation();
  const theme = useTheme();

  // TODO: Implement
  const { mutate: replayEvent } = useReplayEvent();

  const onSubmit: SubmitHandler<NewEventSchemaType> = () => {
    replayEvent(event.id);
  };

  const { handleSubmit } = useForm<NewEventSchemaType>({
    resolver: zodResolver(NewEventSchema),
    values: {
      body: body || "",
    },
  });

  return (
    <form
      id="send-event"
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col space-y-5"
    >
      <DialogHeader>
        <DialogTitle>
          <Calendar />
          {t("pages.events.history.view.dialogHeader")}
        </DialogTitle>
      </DialogHeader>
      <Card
        className="grow p-4 pl-0"
        background="weight-1"
        data-testid="variable-create-card"
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
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button data-testid="event-view-submit" type="submit" variant="primary">
          {t("pages.events.history.view.submit")}
        </Button>
      </DialogFooter>
    </form>
  );
};

export default EventEditorForm;
