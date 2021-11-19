import React from 'react';
import SecretsPanel from './secrets-panel';
import RegistriesPanel from './registries-panel';
import FlexBox from '../../components/flexbox';
import BroadcastConfigurationsPanel from './broadcast-panel';
import VariablesPanel from './variables-panel';
import ScarySettings from './scary-panel';

function ExamplePage(props) {
    return(
        <FlexBox id="settings-page" className="col gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="gap col">
                <FlexBox className="gap wrap">
                    <FlexBox style={{ minWidth: "380px" }}>
                        <SecretsPanel />
                    </FlexBox>
                    <FlexBox style={{ minWidth: "380px" }}>
                        <RegistriesPanel />
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
                <ScarySettings />
            </FlexBox>
        </FlexBox>
    )
}

export default ExamplePage;