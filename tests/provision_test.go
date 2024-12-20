/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("[zfspv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with zfs driver", func() {
		It("Running zfs volume Creation Test", volumeCreationTest)
		It("Running zfs volume Creation Test with custom node id", Label("custom-node-id"), volumeCreationTest)
		It("Running zfs volume Deletion Test", volumeDeletionTest)
	})
})

func fsVolCreationTest() {
	storageClass := getStoragClassParams()
	for _, params := range storageClass {
		exhaustiveVolumeTests(params)
	}
}

// Test to cater create, snapshot, clone and delete resources
func exhaustiveVolumeTests(parameters map[string]string) {
	fstype := parameters["fstype"]
	create(parameters)
	snapshotAndCloneCreate()
	// btrfs does not support online resize
	if fstype != "btrfs" {
		By("Resizing the PVC", func() { resizeAndVerifyPVC(pvcNameFS) })
	}
	snapshotAndCloneCleanUp()
	cleanUp()
}

// Creates the resources
func create(parameters map[string]string) {
	By("####### Creating the storage class : " + parameters["fstype"] + " #######")
	createFstypeStorageClass(parameters)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameFS) })
	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameFS, pvcNameFS) })
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying storage class parameters")
	VerifyStorageClassParams(parameters)
}

// Creates the snapshot/clone resources
func snapshotAndCloneCreate() {
	createSnapshot(pvcNameFS, snapNameFS)
	verifySnapshotCreated(snapNameFS)
	createClone(clonePvcNameFS, snapNameFS, scObj.Name)
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameFS, clonePvcNameFS) })
}

// Removes the snapshot/clone resources
func snapshotAndCloneCleanUp() {
	deleteAppDeployment(cloneAppNameFS)
	deletePVC(clonePvcNameFS)
	deleteSnapshot(pvcNameFS, snapNameFS)
}

// Removes the resources
func cleanUp() {
	deleteAppDeployment(appNameFS)
	deletePVC(pvcNameFS)
	By("Deleting storage class", deleteStorageClass)
}

func blockVolCreationTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameBlock) })

	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameBlock, pvcNameBlock) })
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying ZFSVolume property change", VerifyZFSVolumePropEdit)

	createSnapshot(pvcNameBlock, snapNameBlock)
	verifySnapshotCreated(snapNameBlock)
	createClone(clonePvcNameBlock, snapNameBlock, scObj.Name)
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameBlock, clonePvcNameBlock) })

	By("Deleting clone and main application deployment")
	deleteAppDeployment(cloneAppNameBlock)
	deleteAppDeployment(appNameBlock)

	By("Deleting snapshot, main pvc and clone pvc")
	deletePVC(clonePvcNameBlock)
	deleteSnapshot(pvcNameBlock, snapNameBlock)
	deletePVC(pvcNameBlock)

	By("Deleting storage class", deleteStorageClass)
}

func blockVolDeletionTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameAplha) })

	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameAlpha, pvcNameAplha) })
	By("verifying ZFSVolume object", VerifyZFSVolume)

	createSnapshot(pvcNameAplha, snapNameAlpha)
	verifySnapshotCreated(snapNameAlpha)

	By("Deleting main application deployment")
	deleteAppDeployment(appNameAlpha)

	By("Deleting main pvc")
	deletePVC(pvcNameAplha)

	By("Verifying ZFSVolume object after pvc deletion when snapshot is present", VerifyZFSVolume)

	By("Creating clone from the snapshot")
	createClone(clonePvcNameAlpha, snapNameAlpha, scObj.Name)
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameAlpha, clonePvcNameAlpha) })

	By("Deleting clone application deployment, clone pvc")
	deleteAppDeployment(cloneAppNameAlpha)

	deletePVC(clonePvcNameAlpha)
	deleteSnapshot(pvcNameAplha, snapNameAlpha)

	By("Deleting storage class", deleteStorageClass)
}

func volumeCreationTest() {
	By("Running volume creation test", fsVolCreationTest)
	By("Running block volume creation test", blockVolCreationTest)

}

func volumeDeletionTest() {
	By("Running volume deletion test", blockVolDeletionTest)
}
