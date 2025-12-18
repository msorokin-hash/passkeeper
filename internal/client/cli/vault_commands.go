package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	"github.com/spf13/cobra"
)

// DeleteDataCmd returns a Cobra command that deletes a vault item by its ID.
// The command requires the --id flag.
func (c *CLI) DeleteDataCmd(ctx context.Context) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete",
		Long:  "Delete data by ID",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.service.DeleteData(ctx, id)
			if err != nil {
				return err
			}
			fmt.Println("Data successfully deleted")
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "ID of the data to delete")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// AddFileCmd returns a Cobra command that stores a file in the vault.
// The command requires the --path flag and accepts an optional --comment.
func (c *CLI) AddFileCmd(ctx context.Context) *cobra.Command {
	var file, comment string
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Add file",
		Long:  "Add a file to the vault",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.service.AddFile(ctx, file, comment)
			if err != nil {
				return err
			}
			fmt.Println("File successfully added to the vault")
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "path", "p", "", "file path")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().StringVarP(&comment, "comment", "c", "", "comment")

	return cmd
}

// AddBankCardCmd returns a Cobra command that stores bank card data in the vault.
// The command requires card holder, number, CSV, expiry month/year, and bank,
// and accepts an optional comment.
func (c *CLI) AddBankCardCmd(ctx context.Context) *cobra.Command {
	var cardData entity.BankCardData
	var bankName, comment string

	cmd := &cobra.Command{
		Use:   "bcard",
		Short: "Add bank card",
		Long:  "Add bank card information to the vault",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.service.AddBankCard(ctx, cardData, entity.BankCardMeta{
				Bank:    bankName,
				Comment: comment,
			})
			if err != nil {
				return err
			}
			fmt.Println("Bank card successfully added to the vault")
			return nil
		},
	}

	cmd.Flags().StringVarP(&cardData.Holder, "owner", "o", "", "card holder")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringVarP(&cardData.Number, "number", "n", "", "card number")
	_ = cmd.MarkFlagRequired("number")
	cmd.Flags().StringVarP(&cardData.CSV, "csv", "s", "", "CSV code")
	_ = cmd.MarkFlagRequired("csv")
	cmd.Flags().IntVarP(&cardData.ValidMonth, "month", "m", 0, "valid until month")
	_ = cmd.MarkFlagRequired("month")
	cmd.Flags().IntVarP(&cardData.ValidYear, "year", "y", 0, "valid until year")
	_ = cmd.MarkFlagRequired("year")
	cmd.Flags().StringVarP(&bankName, "bank", "b", "", "bank name")
	_ = cmd.MarkFlagRequired("bank")
	cmd.Flags().StringVarP(&comment, "comment", "c", "", "comment")

	return cmd
}

// AddTextCmd returns a Cobra command that stores a text entry in the vault.
// The command requires the text body and name via flags and accepts an optional comment.
func (c *CLI) AddTextCmd(ctx context.Context) *cobra.Command {
	var data, name, comment string

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Add text",
		Long:  "Add text information to the vault",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.service.AddText(ctx, data, entity.TextMeta{
				Name:    name,
				Comment: comment,
			})
			if err != nil {
				return err
			}
			fmt.Println("Text successfully added to the vault")
			return nil
		},
	}
	cmd.Flags().StringVarP(&data, "text", "t", "", "text content")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().StringVarP(&name, "name", "n", "", "text name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&comment, "comment", "c", "", "comment")

	return cmd
}

// AddPasswordCmd returns a Cobra command that stores a password entry in the vault.
// The command requires password, login, and resource via flags and accepts an optional comment.
func (c *CLI) AddPasswordCmd(ctx context.Context) *cobra.Command {
	var password, login, resource, comment string

	cmd := &cobra.Command{
		Use:   "password",
		Short: "Add password",
		Long:  "Add a new password to the vault",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.service.AddPassword(ctx, password, entity.PasswordMeta{
				Resource: resource,
				Login:    login,
				Comment:  comment,
			})
			if err != nil {
				return err
			}
			fmt.Println("Password successfully added to the vault")
			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "password to store")
	_ = cmd.MarkFlagRequired("password")
	cmd.Flags().StringVarP(&login, "login", "l", "", "login for the resource")
	_ = cmd.MarkFlagRequired("login")
	cmd.Flags().StringVarP(&resource, "resource", "r", "", "resource name")
	_ = cmd.MarkFlagRequired("resource")
	cmd.Flags().StringVarP(&comment, "comment", "c", "", "comment")

	return cmd
}

// AddDataCmd returns a parent Cobra command that groups subcommands for
// adding different types of data (passwords, text, bank cards, files) to the vault.
func (c *CLI) AddDataCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add",
		Long:  "Upload data into the vault",
	}

	cmd.AddCommand(
		c.AddPasswordCmd(ctx),
		c.AddTextCmd(ctx),
		c.AddBankCardCmd(ctx),
		c.AddFileCmd(ctx),
	)

	return cmd
}

// GetDataCmd returns a Cobra command that retrieves and prints a single vault item by ID.
// The command requires the --id flag and prints a formatted representation
// depending on the item's data type.
func (c *CLI) GetDataCmd(ctx context.Context) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get data from vault",
		Long:  "Retrieve data from the vault by ID",
		RunE: func(_ *cobra.Command, _ []string) error {
			dataType, res, err := c.service.GetData(ctx, id)
			if err != nil {
				return err
			}

			switch dataType {
			case entity.Password:
				item, ok := res.(*entity.PasswordItem)
				if !ok {
					return errors.New("invalid password data")
				}
				fmt.Printf(
					"Resource: %s; Login: %s; Password: %s\nComment: %s\n",
					item.Meta.Resource,
					item.Meta.Login,
					item.Data,
					item.Meta.Comment,
				)
				return nil

			case entity.BankCard:
				item, ok := res.(*entity.BankCardItem)
				if !ok {
					return errors.New("invalid bank card data")
				}
				fmt.Printf(
					"Bank: %s\nHolder: %s; Number: %s; Valid until: %d-%d; CSV: %s\nComment: %s\n",
					item.Meta.Bank,
					item.Data.Holder,
					item.Data.Number,
					item.Data.ValidMonth,
					item.Data.ValidYear,
					item.Data.CSV,
					item.Meta.Comment,
				)
				return nil

			case entity.Text:
				item, ok := res.(*entity.TextItem)
				if !ok {
					return errors.New("invalid text data")
				}
				fmt.Printf(
					"Name: %s\nText: %s\nComment: %s\n",
					item.Meta.Name, item.Data, item.Meta.Comment,
				)
				return nil

			case entity.File:
				item, ok := res.(*entity.FileItem)
				if !ok {
					return errors.New("invalid file data")
				}
				fmt.Printf(
					"File '%s' downloaded successfully.\n",
					item.Meta.Name,
				)
				return nil
			}

			return fmt.Errorf("unknown data type: %s", dataType)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "ID of the data to retrieve")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// GetAllByTypeCmd returns a Cobra command that lists stored items of a given data type.
// The command requires the --type flag and prints metadata for each matching item.
func (c *CLI) GetAllByTypeCmd(ctx context.Context) *cobra.Command {
	var dataType string
	cmd := &cobra.Command{
		Use:   "getall",
		Short: "List data",
		Long:  "Retrieve a list of stored data items of a specific type",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := c.service.GetAllByType(ctx, entity.DataType(dataType))
			if err != nil {
				return err
			}
			if len(res) == 0 {
				fmt.Println("No data found")
				return nil
			}

			switch dataType {
			case string(entity.Password):
				for _, item := range res {
					meta, ok := item.Meta.(*entity.PasswordMeta)
					if !ok {
						return errors.New("unable to read password metadata")
					}
					fmt.Printf(
						"id: %s; Resource: %s; Login: %s; Comment: %s\n",
						item.ID, meta.Resource, meta.Login, meta.Comment,
					)
				}

			case string(entity.Text):
				for _, item := range res {
					meta, ok := item.Meta.(*entity.TextMeta)
					if !ok {
						return errors.New("unable to read text metadata")
					}
					fmt.Printf(
						"id: %s; Name: %s; Comment: %s\n",
						item.ID, meta.Name, meta.Comment,
					)
				}

			case string(entity.BankCard):
				for _, item := range res {
					meta, ok := item.Meta.(*entity.BankCardMeta)
					if !ok {
						return errors.New("unable to read bank card metadata")
					}
					fmt.Printf(
						"id: %s; Bank: %s; Comment: %s\n",
						item.ID, meta.Bank, meta.Comment,
					)
				}

			case string(entity.File):
				for _, item := range res {
					meta, ok := item.Meta.(*entity.FileMeta)
					if !ok {
						return errors.New("unable to read file metadata")
					}
					fmt.Printf(
						"id: %s; Name: %s; Extension: %s; Comment: %s\n",
						item.ID, meta.Name, meta.Extension, meta.Comment,
					)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataType, "type", "t", "", "data type for listing")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
