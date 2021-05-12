'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        for (let i=0; i<this.roundArguments.assets; i++) {
            const assetID = `${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Creating asset ${assetID}`);
            const request = {
                contractId: this.roundArguments.contractId, 
                contractFunction: 'Invoke',
                invokerIdentity: 'User1',
                contractArguments: ["createTokenizeAsset", assetID, "75000", "16000", "USD-T", "1", "1DPT", "45000","Dubai"],
                readOnly: false
            }; //contractFunction: 'CreateAsset',  contractArguments: [assetID,'blue','20','penguin','500'],

            await this.sutAdapter.sendRequests(request);
        }
    }
    
    async submitTransaction() {
        const randomId = Math.floor(Math.random()*this.roundArguments.assets);
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'Invoke',
            invokerIdentity: 'User1',
            contractArguments: ["createTokenizeAsset", "KZMall", "75000", "16000", "USD-T", "1", "1DPT", "45000","Dubai", 
            `${this.workerIndex}_${randomId}`],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(myArgs);
    }
    
    async cleanupWorkloadModule() {
        for (let i=0; i<this.roundArguments.assets; i++) {
            const assetID = `${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Deleting asset ${assetID}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'Invoke',
                invokerIdentity: 'User1',
                contractArguments: ["deleteTokenizeAsset",assetID],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
}

function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
