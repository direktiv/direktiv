/* eslint-disable tailwindcss/no-custom-classname */
import "./style.css";

import { Config, GenerateRandomKey } from "../../../util";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../../components/content-panel";
import React, { useState } from "react";
import { VscAdd, VscLock, VscTrash } from "react-icons/vsc";

import Alert from "../../../components/alert";
import { AutoSizer } from "react-virtualized";
import DirektivEditor from "../../../components/editor";
import FlexBox from "../../../components/flexbox";
import HelpIcon from "../../../components/help";
import Modal from "../../../components/modal";
import Tabs from "../../../components/tabs";
import { useApiKey } from "../../../util/apiKeyProvider";
import { useDropzone } from "react-dropzone";
import { useSecrets } from "../../../hooks";

function SecretsPanel(props) {
  const { namespace } = props;

  const [keyValue, setKeyValue] = useState("");
  const [file, setFile] = useState(null);
  const [vValue, setVValue] = useState("");
  const [apiKey] = useApiKey();
  const { data, createSecret, deleteSecret, getSecrets } = useSecrets(
    Config.url,
    namespace,
    apiKey
  );

  return (
    <ContentPanel style={{ height: "100%", minHeight: "180px", width: "100%" }}>
      <ContentPanelTitle>
        <ContentPanelTitleIcon>
          <VscLock />
        </ContentPanelTitleIcon>
        <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
          <div>Secrets</div>
          <HelpIcon msg="Encrypted key/value pairs that can be referenced within workflows. Suitable for storing sensitive information (such as tokens) for use in workflows." />
        </FlexBox>
        <div>
          <Modal
            title="New secret"
            escapeToCancel
            titleIcon={<VscLock />}
            modalStyle={{ width: "600px" }}
            onClose={() => {
              setKeyValue("");
              setVValue("");
              setFile(null);
            }}
            button={<VscAdd />}
            buttonProps={{
              auto: true,
            }}
            actionButtons={[
              {
                label: "Add",

                onClick: async () => {
                  if (document.getElementById("file-picker")) {
                    if (keyValue.trim() === "") {
                      throw new Error("Secret key name needs to be provided.");
                    }
                    if (!file) {
                      throw new Error("Please add or select file");
                    }
                    await createSecret(keyValue, file);
                  } else {
                    if (keyValue.trim() === "") {
                      throw new Error("Secret key name needs to be provided.");
                    }
                    if (vValue.trim() === "") {
                      throw new Error("Secret value needs to be provided.");
                    }
                    await createSecret(keyValue, vValue);
                  }
                  await getSecrets();
                },

                buttonProps: { variant: "contained", color: "primary" },
                closesModal: true,
                validate: true,
              },
              {
                label: "Cancel",
                closesModal: true,
              },
            ]}
            requiredFields={[
              { tip: "secret key is required", value: keyValue },
            ]}
          >
            <Tabs
              style={{ minHeight: "100px", minWidth: "400px" }}
              headers={["Manual", "Upload"]}
              tabs={[
                // no key needed for static array
                // eslint-disable-next-line react/jsx-key
                <AddSecretPanel
                  keyValue={keyValue}
                  vValue={vValue}
                  setKeyValue={setKeyValue}
                  setVValue={setVValue}
                />,
                // no key needed for static array
                // eslint-disable-next-line react/jsx-key
                <FlexBox
                  id="file-picker"
                  className="col gap"
                  style={{ fontSize: "12px" }}
                >
                  <div
                    style={{
                      width: "100%",
                      paddingRight: "12px",
                      display: "flex",
                    }}
                  >
                    <input
                      value={keyValue}
                      onChange={(e) => setKeyValue(e.target.value)}
                      autoFocus
                      placeholder="Enter key"
                    />
                  </div>
                  <FlexBox id="file-picker" gap>
                    <SecretFilePicker
                      file={file}
                      setFile={setFile}
                      id="add-secret-panel"
                    />
                  </FlexBox>
                </FlexBox>,
              ]}
            />
          </Modal>
        </div>
      </ContentPanelTitle>
      <ContentPanelBody className="secrets-panel">
        <FlexBox col gap>
          <FlexBox className="secrets-list">
            {data !== null ? (
              <Secrets
                deleteSecret={deleteSecret}
                getSecrets={getSecrets}
                secrets={data}
              />
            ) : (
              ""
            )}
          </FlexBox>
          <div>
            <Alert severity="info">
              Once a secret is removed, it can never be restored.
            </Alert>
          </div>
        </FlexBox>
      </ContentPanelBody>
    </ContentPanel>
  );
}

export default SecretsPanel;

export function SecretFilePicker(props) {
  const { file, setFile, id } = props;

  const onDrop = (acceptedFiles) => {
    setFile(acceptedFiles[0]);
  };

  const { getRootProps, getInputProps } = useDropzone({
    onDrop,
    multiple: false,
  });

  return (
    <div
      {...getRootProps()}
      className="file-input"
      id={id}
      style={{ display: "flex", flex: "auto", flexDirection: "column" }}
    >
      <div>
        <input {...getInputProps()} />
        <p>Drag &apos;n&apos; drop the file here, or click to select file</p>
        {file !== null ? (
          <p style={{ margin: "0px" }}>
            Selected file: &apos;{file.path}&apos;
          </p>
        ) : (
          ""
        )}
      </div>
    </div>
  );
}

function Secrets(props) {
  const { secrets, deleteSecret, getSecrets } = props;

  return (
    <>
      <FlexBox col gap style={{ maxHeight: "236px", overflowY: "auto" }}>
        {secrets.length === 0 ? (
          <FlexBox className="secret-tuple empty-content">
            <FlexBox className="key">No secrets are stored...</FlexBox>
            <FlexBox className="val"></FlexBox>
            <FlexBox className="actions"></FlexBox>
          </FlexBox>
        ) : (
          <>
            {secrets.map((obj) => {
              const key = GenerateRandomKey("secret-");

              return (
                <FlexBox className="secret-tuple" key={key} id={key}>
                  <FlexBox className="key">{obj.name}</FlexBox>
                  <FlexBox className="val">
                    <span>******</span>
                  </FlexBox>
                  <FlexBox className="actions">
                    <Modal
                      modalStyle={{ width: "360px" }}
                      escapeToCancel
                      style={{
                        flexDirection: "row-reverse",
                        marginRight: "8px",
                      }}
                      titleIcon={<VscLock />}
                      title="Remove secret"
                      button={<SecretsDeleteButton />}
                      buttonProps={{
                        variant: "text",
                        color: "info",
                      }}
                      actionButtons={[
                        {
                          label: "Delete",

                          onClick: async () => {
                            await deleteSecret(obj.name);
                            await getSecrets();
                          },
                          buttonProps: { variant: "contained", color: "error" },
                          closesModal: true,
                        },
                        {
                          label: "Cancel",
                          closesModal: true,
                        },
                      ]}
                    >
                      <FlexBox col gap>
                        <FlexBox>
                          Are you sure you want to delete &apos;{obj.name}
                          &apos;?
                          <br />
                          This action cannot be undone.
                        </FlexBox>
                      </FlexBox>
                    </Modal>
                  </FlexBox>
                </FlexBox>
              );
            })}
          </>
        )}
      </FlexBox>
    </>
  );
}

export function SecretsDeleteButton() {
  return (
    <div
      className="red-text"
      style={{ display: "flex", alignItems: "center", height: "100%" }}
    >
      <VscTrash />
    </div>
  );
}

function AddSecretPanel(props) {
  const { keyValue, vValue, setKeyValue, setVValue } = props;

  return (
    <FlexBox col gap style={{ fontSize: "12px", width: "400px" }}>
      <FlexBox gap>
        <FlexBox>
          <input
            value={keyValue}
            onChange={(e) => setKeyValue(e.target.value)}
            autoFocus
            placeholder="Enter key"
          />
        </FlexBox>
      </FlexBox>
      <FlexBox gap style={{ minHeight: "250px" }}>
        <FlexBox style={{ overflow: "hidden" }}>
          <AutoSizer>
            {({ height, width }) => (
              <DirektivEditor
                width={width}
                dvalue={vValue}
                setDValue={setVValue}
                height={height}
              />
            )}
          </AutoSizer>
        </FlexBox>
      </FlexBox>
    </FlexBox>
  );
}
