'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class CreateCampaignWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
        this.ids = [];
    }

    /**
     * Assemble TXs for the round.
     * @return {Promise<TxStatus[]>}
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
    }

    async submitTransaction() {
        this.txIndex++;
        const randID = Math.floor(Math.random() * 1000)
        const assetID = randID.toString() + `_${this.workerIndex}_${this.txIndex}`;
        this.ids.push(assetID)

        console.log(`Worker ${this.workerIndex}: Creating asset ${assetID}`);
        const request = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'CreateCampaign',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [assetID, 'Camp1', '"2022-05-02T15:02:40.628Z"', '"2023-05-02T15:02:40.628Z"'],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
    async cleanupWorkloadModule() { }
}

/**
 * Create a new instance of the workload module.
 * @return {WorkloadModuleInterface}
 */

function createWorkloadModule() {
    return new CreateCampaignWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;