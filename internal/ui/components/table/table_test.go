package table

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTable_Render_Basic(t *testing.T) {
	columns := []Column{
		{Key: "name", Label: "Name"},
		{Key: "status", Label: "Status"},
	}
	rows := []map[string]interface{}{
		{"name": "Item 1", "status": "Active"},
		{"name": "Item 2", "status": "Inactive"},
	}

	props := TableProps{
		Columns: columns,
		Rows:    rows,
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Name")
	assert.Contains(t, content, "Status")
	assert.Contains(t, content, "Item 1")
	assert.Contains(t, content, "Active")
	assert.Contains(t, content, "<table")
}

func TestTable_Render_WithSorting(t *testing.T) {
	columns := []Column{
		{Key: "name", Label: "Name", Sortable: true},
		{Key: "date", Label: "Date", Sortable: true},
	}
	rows := []map[string]interface{}{
		{"name": "Alpha", "date": "2024-01-01"},
	}

	props := TableProps{
		Columns:   columns,
		Rows:      rows,
		Sortable:  true,
		SortBy:    "name",
		SortOrder: "asc",
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "Alpha")
}

func TestTable_Render_Striped(t *testing.T) {
	columns := []Column{{Key: "id", Label: "ID"}}
	rows := []map[string]interface{}{
		{"id": "1"},
		{"id": "2"},
	}

	props := TableProps{
		Columns: columns,
		Rows:    rows,
		Striped: true,
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "1")
}

func TestTable_Render_Hoverable(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []map[string]interface{}{
		{"name": "Test"},
	}

	props := TableProps{
		Columns:   columns,
		Rows:      rows,
		Hoverable: true,
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Test")
}

func TestTable_Render_WithPagination(t *testing.T) {
	columns := []Column{{Key: "id", Label: "ID"}}
	rows := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		rows[i] = map[string]interface{}{"id": i + 1}
	}

	props := TableProps{
		Columns:     columns,
		Rows:        rows,
		Paginated:   true,
		PageSize:    10,
		CurrentPage: 1,
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "1")
	assert.Contains(t, content, "10")
}

func TestTable_Render_Compact(t *testing.T) {
	columns := []Column{
		{Key: "name", Label: "Name"},
		{Key: "value", Label: "Value"},
	}
	rows := []map[string]interface{}{
		{"name": "Compact Row", "value": "123"},
	}

	props := TableProps{
		Columns: columns,
		Rows:    rows,
		Compact: true,
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Compact Row")
}

func TestTable_Render_Empty(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}

	props := TableProps{
		Columns:    columns,
		Rows:       []map[string]interface{}{},
		EmptyState: "No data available",
	}

	var buf bytes.Buffer
	err := Table(props).Render(context.Background(), &buf)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No data available")
}
