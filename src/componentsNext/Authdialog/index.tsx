import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { KeyRound, LogIn } from "lucide-react";

import Button from "~/design/Button";
import Input from "~/design/Input";
import Logo from "~/design/Logo";
import { useTranslation } from "react-i18next";

export const Authdialog = () => {
  const { t } = useTranslation();

  const isLoading = false; // TODO:

  // TODO: check if there is an old, and show the user if its wrong
  // TODO: delete old one if not needed anymore

  const formId = `authdialog-form`;

  return (
    <Dialog open={true}>
      <DialogContent className="max-sm:top-20">
        <div className="absolute -top-14 flex w-full justify-center">
          <Logo />
        </div>
        <DialogHeader>
          <DialogTitle>
            <KeyRound /> {t("pages.authenticate.title")}
          </DialogTitle>
        </DialogHeader>
        <div className="my-3">
          {/* <FormErrors errors={errors} className="mb-5" /> */}
          {/* <form id={formId} onSubmit={handleSubmit(onSubmit)}> */}
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.authenticate.apiKey")}
            </label>
            <Input
              id="name"
              placeholder={t("pages.authenticate.apiKeyPlaceholder")}
              // {...register("name")}
            />
          </fieldset>
          {/* </form> */}
        </div>
        <DialogFooter>
          <Button
          // type="submit"
          // disabled={disableSubmit}
          // loading={isLoading}
          // form={formId}
          >
            {!isLoading && <LogIn />}
            {t("pages.authenticate.loginBtn")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
