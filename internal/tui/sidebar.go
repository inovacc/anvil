package tui

import (
	"strings"
)

const sidebarWidth = 22

type sidebarCategory struct {
	label string
	items []sidebarItem
}

type sidebarItem struct {
	label  string
	screen screen
}

type sidebarModel struct {
	categories []sidebarCategory
	cursor     int // flat index across all selectable items
	focused    bool
}

func newSidebarModel() sidebarModel {
	return sidebarModel{
		categories: []sidebarCategory{
			{
				label: "Secrets",
				items: []sidebarItem{
					{"Profiles", screenProfiles},
					{"Secrets", screenSecrets},
					{"Import/Export", screenImportExport},
				},
			},
			{
				label: "Security",
				items: []sidebarItem{
					{"Password", screenPassword},
					{"Key Rotation", screenKeyRotation},
					{"Seal/Unseal", screenSeal},
				},
			},
			{
				label: "Environment",
				items: []sidebarItem{
					{"Release", screenEnvRelease},
					{"Status/Revoke", screenEnvStatus},
					{"Export", screenEnvExport},
				},
			},
			{
				label: "Templates",
				items: []sidebarItem{
					{"List", screenTemplates},
					{"Apply", screenTemplateApply},
				},
			},
			{
				label: "Audit",
				items: []sidebarItem{
					{"Audit Log", screenAudit},
					{"History", screenHistory},
				},
			},
			{
				label: "Integration",
				items: []sidebarItem{
					{"Docker", screenDocker},
					{"Share", screenShare},
					{"Plugins", screenPlugins},
					{"Gather", screenGather},
				},
			},
			{
				label: "Admin",
				items: []sidebarItem{
					{"Status", screenStatus},
					{"Backup/Restore", screenBackup},
				},
			},
		},
		focused: true,
	}
}

func (s sidebarModel) totalItems() int {
	n := 0
	for _, cat := range s.categories {
		n += len(cat.items)
	}
	return n
}

func (s sidebarModel) selectedScreen() screen {
	idx := 0
	for _, cat := range s.categories {
		for _, item := range cat.items {
			if idx == s.cursor {
				return item.screen
			}
			idx++
		}
	}
	return screenStatus
}

// moveCursorToScreen sets the cursor to match the given screen.
func (s *sidebarModel) moveCursorToScreen(sc screen) {
	idx := 0
	for _, cat := range s.categories {
		for _, item := range cat.items {
			if item.screen == sc {
				s.cursor = idx
				return
			}
			idx++
		}
	}
}

func (s *sidebarModel) moveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

func (s *sidebarModel) moveDown() {
	if s.cursor < s.totalItems()-1 {
		s.cursor++
	}
}

func (s sidebarModel) View(height int) string {
	var b strings.Builder
	idx := 0

	for i, cat := range s.categories {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(sidebarCategoryStyle.Render(cat.label))
		b.WriteString("\n")

		for _, item := range cat.items {
			label := "  " + item.label
			if idx == s.cursor {
				if s.focused {
					b.WriteString(sidebarSelectedFocusedStyle.Width(sidebarWidth - 1).Render(label))
				} else {
					b.WriteString(sidebarSelectedStyle.Width(sidebarWidth - 1).Render(label))
				}
			} else {
				b.WriteString(sidebarItemStyle.Width(sidebarWidth - 1).Render(label))
			}
			b.WriteString("\n")
			idx++
		}
	}

	// Pad remaining height
	lines := strings.Count(b.String(), "\n")
	for i := lines; i < height-1; i++ {
		b.WriteString("\n")
	}

	return sidebarStyle.Height(height).Render(b.String())
}
