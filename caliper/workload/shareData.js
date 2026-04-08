'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
let ids = []

class ShareDataWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
        this.campaignID = Math.floor(Math.random() * 1000).toString()
    }

    /**
     * Assemble TXs for the round.
     * @return {Promise<TxStatus[]>}
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        console.log(`Worker ${this.workerIndex}: Creating the campaign ${this.campaignID}`);

        const createCampaign = {
            contractId: "campaign",
            contractFunction: 'CreateCampaign',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [this.campaignID, 'Camp1', '"2022-05-02T15:02:40.628Z"', '"2023-05-02T15:02:40.628Z"'],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(createCampaign);
    }

    async submitTransaction() {
        this.txIndex++;
        const randID = Math.floor(Math.random() * 1000)
        const assetID = randID.toString() + `_${this.workerIndex}_${this.txIndex}`;
        ids.push(assetID)

        console.log(`Worker ${this.workerIndex}: Share KG - Creating asset ${assetID}`);
        const shareKG = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'ShareData',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [assetID, this.campaignID, "abc", "10"],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(shareKG);
    }

    async cleanupWorkloadModule() { }
}

/**
 * Create a new instance of the workload module.
 * @return {WorkloadModuleInterface}
 */

function createWorkloadModule() {
    return new ShareDataWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;