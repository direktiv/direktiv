/* eslint-disable tailwindcss/no-custom-classname */
import "./style.css";

import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../../components/content-panel";
import React, { useEffect, useState } from "react";
import { VscSettingsGear, VscTrash } from "react-icons/vsc";

import Alert from "../../../components/alert";
import FlexBox from "../../../components/flexbox";
import Modal from "../../../components/modal";

function ScarySettings(props) {
  const { deleteNamespace, namespace, deleteErr } = props;
  return (
    <>
      <div>
        <ContentPanel className="scary-panel">
          <ContentPanelTitle>
            <ContentPanelTitleIcon>
              <VscSettingsGear className="red-text" />
            </ContentPanelTitleIcon>
            <FlexBox className="red-text">Important Settings</FlexBox>
          </ContentPanelTitle>
          <ContentPanelBody className="secrets-panel">
            <FlexBox col gap>
              <FlexBox className="scary-settings">
                <Scary
                  namespace={namespace}
                  deleteErr={deleteErr}
                  deleteNamespace={deleteNamespace}
                />
              </FlexBox>
              <FlexBox>
                <Alert severity="error" variant="filled">
                  These settings are super dangerous! Use at your own risk!
                </Alert>
              </FlexBox>
            </FlexBox>
          </ContentPanelBody>
        </ContentPanel>
      </div>
    </>
  );
}

export default ScarySettings;

function Scary(props) {
  const { deleteNamespace, namespace } = props;
  const [delButtonEnabled, setDelButtonEnabled] = useState(false);
  // deleteErr gets filled in when someone attempts to delete a namespace and an error happens

  // let delBtnClasses = "small red";
  // if (!delButtonEnabled) {
  //     delBtnClasses += " disabled"
  // }

  return (
    <>
      <FlexBox>
        <FlexBox
          className="auto-margin"
          style={{ fontSize: "12px", maxWidth: "300px" }}
        >
          This will permanently delete the current active namespace and all
          resources associated with it.
        </FlexBox>
        <FlexBox>
          <Modal
            title="Delete namespace"
            escapeToCancel
            modalStyle={{ width: "360px" }}
            titleIcon={<VscTrash />}
            button={<span>Delete Namespace</span>}
            buttonProps={{
              variant: "contained",
              color: "error",
              tooltip: "Delete Namespace",
              disabledTooltip: "Requires save",
            }}
            requiredFields={[
              {
                tip: "typing namespace name is required",
                value: delButtonEnabled ? "valid" : "",
              },
            ]}
            actionButtons={[
              {
                label: "Delete",
                onClick: async () => {
                  await deleteNamespace(namespace);
                  localStorage.removeItem("namespace");
                  window.location = `/`;
                },
                buttonProps: { variant: "contained", color: "error" },
                closesModal: true,
                validate: true,
              },
              {
                label: "Cancel",
                closesModal: true,
              },
            ]}
          >
            <DeleteNamespaceConfirmationPanel
              namespace={namespace}
              setDelButtonEnabled={setDelButtonEnabled}
            />
          </Modal>
        </FlexBox>
      </FlexBox>
    </>
  );
}

function DeleteNamespaceConfirmationPanel(props) {
  const { namespace, setDelButtonEnabled } = props;

  const [inputValue, setInputValue] = useState("");

  useEffect(() => {
    setDelButtonEnabled(inputValue === namespace);
  }, [inputValue, namespace, setDelButtonEnabled]);

  return (
    <FlexBox col style={{ fontSize: "12px" }}>
      <FlexBox col>
        <p>
          Are you sure you want to delete this namespace?
          <br /> This action <b>can not be undone!</b>
        </p>
        <br />
        <p>
          Please type <b>{namespace}</b> to confirm.
        </p>
      </FlexBox>
      <FlexBox>
        <input
          value={inputValue}
          onChangeCapture={setInputValue}
          onChange={(e) => {
            setInputValue(e.target.value);
            if (e.target.value === namespace) {
              setDelButtonEnabled(true);
            } else {
              setDelButtonEnabled(false);
            }
          }}
          type="text"
        ></input>
      </FlexBox>
    </FlexBox>
  );
}
