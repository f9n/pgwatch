package cmdopts

import (
	"context"
	"errors"
	"fmt"

	"github.com/cybertec-postgresql/pgwatch/v3/internal/sinks"
)

type MaintenanceCommand struct {
	owner                 *Options
	Cleanup               MaintenanceCleanupCommand               `command:"cleanup" description:"Clean up old partitions from metrics database"`
	MaintainUniqueSources MaintenanceMaintainUniqueSourcesCommand `command:"maintain-unique-sources" description:"Update unique sources listings for Grafana dropdowns"`
}

func NewMaintenanceCommand(owner *Options) *MaintenanceCommand {
	return &MaintenanceCommand{
		owner:                 owner,
		Cleanup:               MaintenanceCleanupCommand{owner: owner},
		MaintainUniqueSources: MaintenanceMaintainUniqueSourcesCommand{owner: owner},
	}
}

type MaintenanceCleanupCommand struct {
	owner         *Options
	RetentionDays int  `short:"r" long:"retention-days" description:"Retention period in days for old partitions (overrides global setting)" default:"0"`
	DryRun        bool `short:"n" long:"dry-run" description:"Show what would be deleted without actually deleting"`
}

func (cmd *MaintenanceCleanupCommand) Execute(args []string) error {
	// Initialize sink writer to get access to postgres writer
	err := cmd.owner.InitSinkWriter(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize sink writer: %w", err)
	}

	// Find PostgreSQL sinks and run cleanup
	if _, ok := cmd.owner.SinksWriter.(*sinks.MultiWriter); ok {
		// MultiWriter case - not directly accessible, need to check each sink type
		return errors.New("maintenance cleanup is only supported for single PostgreSQL sinks. Use multiple pgwatch instances or configure a single PostgreSQL sink")
	}

	// Single writer case
	if pgw, ok := cmd.owner.SinksWriter.(*sinks.PostgresWriter); ok {
		retentionDays := cmd.RetentionDays
		if retentionDays == 0 {
			retentionDays = cmd.owner.Sinks.Retention
		}
		if retentionDays <= 0 {
			return errors.New("retention period must be greater than 0")
		}

		return cmd.executeCleanup(pgw, retentionDays)
	}

	return errors.New("maintenance cleanup is only supported for PostgreSQL sinks")
}

func (cmd *MaintenanceCleanupCommand) executeCleanup(pgw *sinks.PostgresWriter, retentionDays int) error {
	dropped, err := pgw.RunPartitionCleanup(retentionDays, cmd.DryRun)
	if err != nil {
		fmt.Printf("Error during partition cleanup: %v\n", err)
		cmd.owner.CompleteCommand(ExitCodeCmdError)
		return nil
	}

	if cmd.DryRun {
		fmt.Printf("DRY RUN: Would delete %d partitions older than %d days\n", dropped, retentionDays)
	} else {
		fmt.Printf("Successfully deleted %d old partitions\n", dropped)
	}

	cmd.owner.CompleteCommand(ExitCodeOK)
	return nil
}

type MaintenanceMaintainUniqueSourcesCommand struct {
	owner *Options
}

func (cmd *MaintenanceMaintainUniqueSourcesCommand) Execute(args []string) error {
	// Initialize sink writer to get access to postgres writer
	err := cmd.owner.InitSinkWriter(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize sink writer: %w", err)
	}

	// Find PostgreSQL sinks and run maintenance
	if _, ok := cmd.owner.SinksWriter.(*sinks.MultiWriter); ok {
		// MultiWriter case - not directly accessible
		return errors.New("maintenance is only supported for single PostgreSQL sinks. Use multiple pgwatch instances or configure a single PostgreSQL sink")
	}

	// Single writer case
	if pgw, ok := cmd.owner.SinksWriter.(*sinks.PostgresWriter); ok {
		return cmd.executeMaintain(pgw)
	}

	return errors.New("maintenance is only supported for PostgreSQL sinks")
}

func (cmd *MaintenanceMaintainUniqueSourcesCommand) executeMaintain(pgw *sinks.PostgresWriter) error {
	err := pgw.RunUniqueSourcesMaintenance()
	if err != nil {
		fmt.Printf("Error during unique sources maintenance: %v\n", err)
		cmd.owner.CompleteCommand(ExitCodeCmdError)
		return nil
	}

	fmt.Printf("Unique sources maintenance completed successfully\n")
	cmd.owner.CompleteCommand(ExitCodeOK)
	return nil
}
