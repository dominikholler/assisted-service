package ocs

import (
	"context"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/operators/api"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/pkg/conversions"
)

var _ = Describe("Ocs Operator", func() {
	var (
		ctx                 = context.TODO()
		operator            = NewOcsOperator(common.GetTestLog())
		masterWithThreeDisk = &models.Host{Role: models.HostRoleMaster,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 32 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}
		masterWithNoDisk      = &models.Host{Role: models.HostRoleMaster, Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 32 * conversions.GiB})}
		masterWithNoInventory = &models.Host{Role: models.HostRoleMaster}
		masterWithOneDisk     = &models.Host{Role: models.HostRoleMaster,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 32 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"}}})}

		masterWithLessCPU = &models.Host{Role: models.HostRoleMaster,
			Inventory: Inventory(&InventoryResources{Cpus: 5, Ram: 32 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}

		masterWithLessRAM = &models.Host{Role: models.HostRoleMaster,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 5 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}
		workerWithOneDisk = &models.Host{Role: models.HostRoleWorker,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 64 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
				}})}
		workerWithTwoDisk = &models.Host{Role: models.HostRoleWorker,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 64 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}
		workerWithThreeDisk = &models.Host{Role: models.HostRoleWorker,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 64 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "HDD"},
				}})}
		workerWithNoDisk      = &models.Host{Role: models.HostRoleWorker, Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 64 * conversions.GiB})}
		workerWithNoInventory = &models.Host{Role: models.HostRoleWorker}
		workerWithLessCPU     = &models.Host{Role: models.HostRoleWorker,
			Inventory: Inventory(&InventoryResources{Cpus: 5, Ram: 64 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}
		workerWithLessRAM = &models.Host{Role: models.HostRoleWorker,
			Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 5 * conversions.GiB,
				Disks: []*models.Disk{
					{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
					{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
				}})}
		autoAssignHost = &models.Host{Role: models.HostRoleAutoAssign, Inventory: Inventory(&InventoryResources{Cpus: 12, Ram: 32 * conversions.GiB,
			Disks: []*models.Disk{
				{SizeBytes: 20 * conversions.GB, DriveType: "HDD"},
				{SizeBytes: 40 * conversions.GB, DriveType: "SSD"},
			}})}
	)

	Context("GetHostRequirements", func() {
		table.DescribeTable("compact mode scenario: get requirements for hosts when ", func(cluster *common.Cluster, host *models.Host, expectedResult *models.ClusterHostRequirementsDetails) {
			res, _ := operator.GetHostRequirements(ctx, cluster, host)
			Expect(res).Should(Equal(expectedResult))
		},
			table.Entry("Single master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk,
				}}},
				masterWithThreeDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + 2*operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + 2*operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are three masters",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithThreeDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + 2*operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + 2*operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("no disk in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithNoDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("no inventory in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoInventory, masterWithOneDisk,
				}}},
				masterWithNoInventory,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("only one disk in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithOneDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are 3 hosts, role of one as auto-assign",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, autoAssignHost,
				}}},
				autoAssignHost,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUCompactMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBCompactMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are two master and one worker",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, workerWithTwoDisk,
				}}},
				workerWithTwoDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
		)

		table.DescribeTable("standard and minimal mode scenario: get requirements for hosts when ", func(cluster *common.Cluster, host *models.Host, expectedResult *models.ClusterHostRequirementsDetails) {
			res, _ := operator.GetHostRequirements(ctx, cluster, host)
			Expect(res).Should(Equal(expectedResult))
		},
			table.Entry("there are 4 hosts, role of one as auto-assign",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, autoAssignHost, masterWithOneDisk,
				}}},
				autoAssignHost,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are 6 hosts, master requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				masterWithThreeDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: 0, RAMMib: 0},
			),
			table.Entry("there are 6 hosts, worker with three disk requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				workerWithThreeDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode + 2*operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode + 2*operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are 6 hosts, worker with two disk requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				workerWithTwoDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode + operator.config.OCSRequiredDiskCPUCount, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode + operator.config.OCSRequiredDiskRAMGiB), DiskSizeGb: MinDiskSize},
			),
			table.Entry("there are 6 hosts, worker with one disk requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithOneDisk,
				}}},
				workerWithOneDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode)},
			),
			table.Entry("there are 6 hosts, worker with no disk requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				workerWithNoDisk,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode)},
			),
			table.Entry("there are 6 hosts, worker with no inventory requirements",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoInventory,
				}}},
				workerWithNoInventory,
				&models.ClusterHostRequirementsDetails{CPUCores: CPUMinimalMode, RAMMib: conversions.GibToMib(MemoryGiBMinimalMode)},
			),
		)
	})

	Context("ValidateHost", func() {
		table.DescribeTable("compact mode scenario: validateHost when ", func(cluster *common.Cluster, host *models.Host, expectedResult api.ValidationResult) {
			res, _ := operator.ValidateHost(ctx, cluster, host)
			Expect(res).Should(Equal(expectedResult))
		}, table.Entry("Single master",
			&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
				masterWithThreeDisk,
			}}},
			masterWithThreeDisk,
			api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
		),
			table.Entry("there are three masters",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithThreeDisk,
				api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
			),
			table.Entry("no disk in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithNoDisk,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient disk to deploy OCS. OCS requires to have at least one non-bootable on each host in compact mode."}},
			),
			table.Entry("only one disk in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk,
				}}},
				masterWithOneDisk,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient disk to deploy OCS. OCS requires to have at least one non-bootable on each host in compact mode."}},
			),
			table.Entry("only one disk in one of the master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithNoInventory,
				}}},
				masterWithNoInventory,
				api.ValidationResult{Status: api.Pending, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Missing Inventory in some of the hosts"}},
			),
			table.Entry("there are 3 hosts, role of one as auto-assign",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, autoAssignHost,
				}}},
				autoAssignHost,
				api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
			),
			table.Entry("there are two master and one worker",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, workerWithTwoDisk,
				}}},
				workerWithTwoDisk,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"OCS unsupported Host Role for Compact Mode."}},
			),
			table.Entry("there are 3 master with less CPU",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithLessCPU,
				}}},
				masterWithLessCPU,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient CPU to deploy OCS. Required CPU count is 8 but found 5."}},
			),
			table.Entry("there are 3 master with less RAM",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithLessRAM,
				}}},
				masterWithLessRAM,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient memory to deploy OCS. Required memory is 24576 MiB but found 5120 MiB."}},
			),
		)

		table.DescribeTable("standard and minimal mode scenario: validateHosts when ", func(cluster *common.Cluster, host *models.Host, expectedResult api.ValidationResult) {
			res, _ := operator.ValidateHost(ctx, cluster, host)
			Expect(res).Should(Equal(expectedResult))
		},
			table.Entry("there are 4 hosts, role of one as auto-assign",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, autoAssignHost, masterWithOneDisk,
				}}},
				autoAssignHost,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"All host roles must be assigned to enable OCS in Standard or Minimal Mode."}},
			),
			table.Entry("there are 6 hosts, master",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				masterWithThreeDisk,
				api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
			),
			table.Entry("there are 6 hosts, worker with two disk",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				workerWithTwoDisk,
				api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
			),
			table.Entry("there are 6 hosts, worker with no disk",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoDisk,
				}}},
				workerWithNoDisk,
				api.ValidationResult{Status: api.Success, ValidationId: operator.GetHostValidationID(), Reasons: []string{}},
			),
			table.Entry("there are 6 hosts, worker with no inventory",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithNoInventory,
				}}},
				workerWithNoInventory,
				api.ValidationResult{Status: api.Pending, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Missing Inventory in some of the hosts"}},
			),
			table.Entry("there are 6 hosts, worker with less CPU",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithLessCPU,
				}}},
				workerWithLessCPU,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient CPU to deploy OCS. Required CPU count is 6 but found 5."}},
			),
			table.Entry("there are 6 hosts, worker with less RAM",
				&common.Cluster{Cluster: models.Cluster{Hosts: []*models.Host{
					masterWithThreeDisk, masterWithNoDisk, masterWithOneDisk, workerWithTwoDisk, workerWithThreeDisk, workerWithLessRAM,
				}}},
				workerWithLessRAM,
				api.ValidationResult{Status: api.Failure, ValidationId: operator.GetHostValidationID(), Reasons: []string{"Insufficient memory to deploy OCS. Required memory is 16384 MiB but found 5120 MiB."}},
			),
		)
	})

})
