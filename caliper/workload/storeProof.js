'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class StoreProofWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.campaignID = Math.floor(Math.random() * 1000).toString();
        this.KGId = Math.floor(Math.random() * 1000).toString();
        this.txIndex = 0;
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

        console.log(`Worker ${this.workerIndex}: Creating anonymized KG ${this.KGId}`);
        const shareKG = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'StoreAnonymizedKG',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [this.KGId, this.campaignID, "rec_id", "env", "sign"],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(shareKG);

    }

    async submitTransaction() {
        this.txIndex++;
        const randID = Math.floor(Math.random() * 1000)
        const randAssetID = randID.toString() + `_${this.workerIndex}_${this.txIndex}`;
        console.log(`Worker ${this.workerIndex}: Verifying the proof for KG ${this.KGId}`);
        const request = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'CaliperStoreProof',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [this.KGId, randAssetID, "a random string of a commitment", "a random string of a commitment"],
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
    return new StoreProofWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;