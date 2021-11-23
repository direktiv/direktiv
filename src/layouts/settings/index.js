import React from 'react';
import SecretsPanel from './secrets-panel';
import RegistriesPanel from './registries-panel';
import FlexBox from '../../components/flexbox';
import BroadcastConfigurationsPanel from './broadcast-panel';
import VariablesPanel from './variables-panel';
import ScarySettings from './scary-panel';

function Settings(props) {
    const {deleteNamespace, namespace} = props

    // if namespace is null top level wait till we have a namespace
    if(namespace === null) {
        return ""
    }

    return(
        <FlexBox id="settings-page" className="col gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="gap col">
                <FlexBox className="gap wrap" style={{ minHeight: "350px" }}>
                    <FlexBox style={{ minWidth: "380px" }}>
                        <SecretsPanel namespace={namespace} />
                    </FlexBox>
                    <FlexBox style={{ minWidth: "380px" }}>
                        <RegistriesPanel namespace={namespace} />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
            <FlexBox>
                <VariablesPanel namespace={namespace} />
            </FlexBox>
            <FlexBox>
                <BroadcastConfigurationsPanel namespace={namespace} />
            </FlexBox>
            <FlexBox>
                <ScarySettings namespace={namespace} deleteNamespace={deleteNamespace} />
            </FlexBox>
        </FlexBox>
    )
}

export default Settings;