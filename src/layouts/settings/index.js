import React from 'react';
import SecretsPanel from './secrets-panel';
import RegistriesPanel from './registries-panel';
import FlexBox from '../../components/flexbox';
import BroadcastConfigurationsPanel from './broadcast-panel';
import VariablesPanel from './variables-panel';
import ScarySettings from './scary-panel';

function Settings(props) {
    const {deleteNamespace, namespace, deleteErr} = props

    // if namespace is null top level wait till we have a namespace
    if(namespace === null) {
        return ""
    }

    return(
        <FlexBox id="settings-page" className="col gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="gap col">
                <FlexBox className="gap wrap">
                    <FlexBox style={{ minWidth: "380px" }}>
                        <SecretsPanel namespace={namespace} />
                    </FlexBox>
                    <FlexBox style={{ minWidth: "380px" }}>
                        <RegistriesPanel namespace={namespace} />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
            <FlexBox>
                <VariablesPanel />
            </FlexBox>
            <FlexBox>
                <BroadcastConfigurationsPanel />
            </FlexBox>
            <FlexBox>
                <ScarySettings deleteErr={deleteErr} namespace={namespace} deleteNamespace={deleteNamespace} />
            </FlexBox>
        </FlexBox>
    )
}

export default Settings;