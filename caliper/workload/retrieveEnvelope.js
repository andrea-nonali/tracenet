'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
let ids = []

class RetreiveEnvelopeWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
        this.campaignID = Math.floor(Math.random() * 1000).toString();
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

        for (let i = 0; i < 30; i++) {
            let KGID = Math.floor(Math.random() * 1000).toString()
            ids.push(KGID)
            console.log(`Worker ${this.workerIndex}: Creating a KG ${KGID}`);

            const shareKG = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'ShareKnowledgeGraph',
                invokerIdentity: 'peer0.obs0.tracenet.com',
                contractArguments: [KGID, this.campaignID, "abc", "10"],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(shareKG);
        }
    }

    async submitTransaction() {
        this.txIndex++;

        var KGID = ids[Math.floor(Math.random() * ids.length)];

        console.log(`Worker ${this.workerIndex}: Query envelope ${KGID}`);
        const queryKG = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'RetrieveEnvelope',
            invokerIdentity: 'peer0.obs0.tracenet.com',
            contractArguments: [KGID],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(queryKG);
    }

    async cleanupWorkloadModule() { }
}

/**
 * Create a new instance of the workload module.
 * @return {WorkloadModuleInterface}
 */

function createWorkloadModule() {
    return new RetreiveEnvelopeWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;