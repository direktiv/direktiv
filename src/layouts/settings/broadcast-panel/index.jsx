/* eslint-disable tailwindcss/no-custom-classname */
import "./style.css";

import { Config, GenerateRandomKey } from "../../../util";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../../components/content-panel";

import Alert from "../../../components/alert";
import FlexBox from "../../../components/flexbox";
import HelpIcon from "../../../components/help";
import React from "react";
import { VscSettings } from "react-icons/vsc";
import { useApiKey } from "../../../util/apiKeyProvider";
import { useBroadcastConfiguration } from "../../../hooks";

function BroadcastConfigurationsPanel(props) {
  const { namespace } = props;
  const [apiKey] = useApiKey();
  const { data, setBroadcastConfiguration, getBroadcastConfiguration } =
    useBroadcastConfiguration(Config.url, namespace, apiKey);

  return (
    <ContentPanel className="broadcast-panel">
      <ContentPanelTitle>
        <ContentPanelTitleIcon>
          <VscSettings />
        </ContentPanelTitleIcon>
        <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
          <div>Broadcast Configurations</div>
          <HelpIcon msg="Toggle which Direktiv system events will cause a Cloud Event to be sent to the current namespace." />
        </FlexBox>
      </ContentPanelTitle>
      {data !== null ? (
        <BroadcastOptions
          getBroadcastConfiguration={getBroadcastConfiguration}
          setBroadcastConfiguration={setBroadcastConfiguration}
          config={data}
        />
      ) : null}
    </ContentPanel>
  );
}

export default BroadcastConfigurationsPanel;

function BroadcastOptions(props) {
  const { config, setBroadcastConfiguration, getBroadcastConfiguration } =
    props;
  const [error, setError] = React.useState(null);
  return (
    <div>
      {error && (
        <ContentPanelBody>
          <Alert severity="warning" variant="standard" grow>
            {error}
          </Alert>
        </ContentPanelBody>
      )}
      <ContentPanelBody>
        <FlexBox>
          <FlexBox col gap>
            <FlexBox className="options-row">
              <BroadcastOptionsRow
                title="Directory"
                options={[
                  {
                    label: "Create",
                    value: config.broadcast["directory.create"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["directory.create"] =
                        !config.broadcast["directory.create"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["directory.create"] =
                          !config.broadcast["directory.create"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Delete",
                    value: config.broadcast["directory.delete"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["directory.delete"] =
                        !config.broadcast["directory.delete"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["directory.delete"] =
                          !config.broadcast["directory.delete"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
              <BroadcastOptionsRow></BroadcastOptionsRow>
            </FlexBox>
            <FlexBox className="options-row">
              <BroadcastOptionsRow
                title="Instance"
                options={[
                  {
                    label: "Success",
                    value: config.broadcast["instance.success"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.success"] =
                        !config.broadcast["instance.success"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.success"] =
                          !config.broadcast["instance.success"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Started",
                    value: config.broadcast["instance.started"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.started"] =
                        !config.broadcast["instance.started"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.started"] =
                          !config.broadcast["instance.started"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Failed",
                    value: config.broadcast["instance.failed"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.failed"] =
                        !config.broadcast["instance.failed"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.failed"] =
                          !config.broadcast["instance.failed"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
              <BroadcastOptionsRow
                title="Instance Variable"
                options={[
                  {
                    label: "Create",
                    value: config.broadcast["instance.variable.create"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.variable.create"] =
                        !config.broadcast["instance.variable.create"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.variable.create"] =
                          !config.broadcast["instance.variable.create"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Update",
                    value: config.broadcast["instance.variable.update"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.variable.update"] =
                        !config.broadcast["instance.variable.update"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.variable.update"] =
                          !config.broadcast["instance.variable.update"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Delete",
                    value: config.broadcast["instance.variable.delete"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["instance.variable.delete"] =
                        !config.broadcast["instance.variable.delete"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["instance.variable.delete"] =
                          !config.broadcast["instance.variable.delete"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
            </FlexBox>
            <FlexBox className="options-row">
              <BroadcastOptionsRow
                title="Namespace Variable"
                options={[
                  {
                    label: "Create",
                    value: config.broadcast["namespace.variable.create"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["namespace.variable.create"] =
                        !config.broadcast["namespace.variable.create"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["namespace.variable.create"] =
                          !config.broadcast["namespace.variable.create"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Update",
                    value: config.broadcast["namespace.variable.update"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["namespace.variable.update"] =
                        !config.broadcast["namespace.variable.update"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["namespace.variable.update"] =
                          !config.broadcast["namespace.variable.update"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Delete",
                    value: config.broadcast["namespace.variable.delete"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["namespace.variable.delete"] =
                        !config.broadcast["namespace.variable.delete"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["namespace.variable.delete"] =
                          !config.broadcast["namespace.variable.delete"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
              <BroadcastOptionsRow />
            </FlexBox>
            <FlexBox>
              <BroadcastOptionsRow
                title="Workflow"
                options={[
                  {
                    label: "Create",
                    value: config.broadcast["workflow.create"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.create"] =
                        !config.broadcast["workflow.create"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.create"] =
                          !config.broadcast["workflow.create"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Update",
                    value: config.broadcast["workflow.update"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.update"] =
                        !config.broadcast["workflow.update"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.update"] =
                          !config.broadcast["workflow.update"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Delete",
                    value: config.broadcast["workflow.delete"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.delete"] =
                        !config.broadcast["workflow.delete"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.delete"] =
                          !config.broadcast["workflow.delete"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
              <BroadcastOptionsRow
                title="Workflow Variable"
                options={[
                  {
                    label: "Create",
                    value: config.broadcast["workflow.variable.create"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.variable.create"] =
                        !config.broadcast["workflow.variable.create"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.variable.create"] =
                          !config.broadcast["workflow.variable.create"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Update",
                    value: config.broadcast["workflow.variable.update"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.variable.update"] =
                        !config.broadcast["workflow.variable.update"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.variable.update"] =
                          !config.broadcast["workflow.variable.update"];
                        setError(error?.message);
                      }
                    },
                  },
                  {
                    label: "Delete",
                    value: config.broadcast["workflow.variable.delete"],
                    onClick: async () => {
                      const cc = config;
                      cc.broadcast["workflow.variable.delete"] =
                        !config.broadcast["workflow.variable.delete"];
                      setError(null);
                      try {
                        await setBroadcastConfiguration(JSON.stringify(cc));
                        await getBroadcastConfiguration();
                      } catch (error) {
                        cc.broadcast["workflow.variable.delete"] =
                          !config.broadcast["workflow.variable.delete"];
                        setError(error?.message);
                      }
                    },
                  },
                ]}
              />
            </FlexBox>
          </FlexBox>
        </FlexBox>
      </ContentPanelBody>
    </div>
  );
}

function BroadcastOptionsRow(props) {
  const { title, options } = props;
  const opts = [];

  for (let i = 0; i < 3; i++) {
    const key = GenerateRandomKey("broadcast-opt-");

    if (!options || i >= options.length) {
      opts.push(
        <FlexBox id={key} key={key} className="col gap">
          <FlexBox></FlexBox>
          <FlexBox key={"broadcast-opts-" + title + "-" + i}>
            <label className="switch" style={{ visibility: "hidden" }}>
              <input type="checkbox" />
              <span className="slider-broadcast"></span>
            </label>
          </FlexBox>
        </FlexBox>
      );
    } else {
      opts.push(
        <FlexBox id={key} key={key} className="col gap broadcast-option">
          <FlexBox>{options[i].label}</FlexBox>
          <FlexBox key={"broadcast-opts-" + title + "-" + i}>
            <label className="switch">
              <input
                onClick={() => {
                  options[i].onClick();
                }}
                defaultChecked={options[i].value}
                type="checkbox"
              />
              <span className="slider-broadcast"></span>
            </label>
          </FlexBox>
        </FlexBox>
      );
    }
  }

  return (
    <FlexBox className="col broadcast-options-panel gap">
      <div className="broadcast-options-header">{title}</div>
      <FlexBox className="broadcast-options-inputs gap">{opts}</FlexBox>
    </FlexBox>
  );
}
