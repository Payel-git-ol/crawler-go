package main

import (
	"github.com/gofiber/fiber/v3"
)

func GetContacts(c fiber.Ctx) error {
	contacts, err := StorageService.GetAllContacts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	result := make([]fiber.Map, 0, len(contacts))

	for _, ct := range contacts {
		result = append(result, fiber.Map{
			"login": ct.Login,
			"hash":  ct.Hash,
		})
	}
	return c.JSON(result)
}

func GetContactsLogin(c fiber.Ctx) error {
	login := c.Params("login")

	contact, err := StorageService.GetContact(login)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
	}

	return c.JSON(fiber.Map{
		"login": contact.Login,
		"hash":  contact.Hash,
	})
}
