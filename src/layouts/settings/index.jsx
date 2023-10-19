import "./style.css";

import BroadcastConfigurationsPanel from "./broadcast-panel";
import FlexBox from "../../components/flexbox";
import React from "react";
import RegistriesPanel from "./registries-panel";
import ScarySettings from "./scary-panel";
import SecretsPanel from "./secrets-panel";
import VariablesPanel from "./variables-panel";

function Settings(props) {
  const { deleteNamespace, namespace } = props;

  // if namespace is null top level wait till we have a namespace
  if (!namespace) {
    return null;
  }

  return (
    <FlexBox
      id="settings-page"
      className="col gap"
      style={{ paddingRight: "8px" }}
    >
      <FlexBox col gap>
        <FlexBox gap wrap style={{ minHeight: "350px" }}>
          <div id="secrets-panel-container">
            <SecretsPanel namespace={namespace} />
          </div>
          <div id="registries-panel-container">
            <RegistriesPanel namespace={namespace} />
          </div>
        </FlexBox>
      </FlexBox>
      <FlexBox>
        <VariablesPanel namespace={namespace} />
      </FlexBox>
      <FlexBox>
        <BroadcastConfigurationsPanel namespace={namespace} />
      </FlexBox>
      <FlexBox>
        <ScarySettings
          namespace={namespace}
          deleteNamespace={deleteNamespace}
        />
      </FlexBox>
    </FlexBox>
  );
}

export default Settings;
