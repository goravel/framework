package filesystem

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

func newFakeTestStorage(t *testing.T) (*Storage, string) {
	root := t.TempDir()
	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("filesystems.default").Return("local").Once()
	config.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
	config.EXPECT().GetString("filesystems.disks.local.root").Return(root).Once()
	config.EXPECT().GetString("filesystems.disks.local.url").Return("https://example.com/storage").Maybe()
	config.EXPECT().GetString("filesystems.disks.s3.url").Return("").Maybe()

	storage, err := NewStorage(config)
	require.NoError(t, err)

	return storage, root
}

func TestFake_ReplacesTheDefaultDisk(t *testing.T) {
	storage, root := newFakeTestStorage(t)
	original := storage.Driver

	fake := storage.Fake()

	assert.Same(t, fake, storage.Disk("local"))
	assert.NoError(t, storage.Put("avatars/1.png", "image"))
	assert.NoError(t, storage.Disk("local").Put("avatars/2.png", "image"))

	assert.NoFileExists(t, root+"/avatars/1.png")
	assert.True(t, fake.AssertExists("avatars/1.png", "avatars/2.png", "avatars"))
	assert.False(t, fake.AssertExists("avatars/1.png", "avatars/3.png"))
	assert.True(t, fake.AssertMissing("avatars/3.png", "other"))
	assert.False(t, fake.AssertMissing("avatars/3.png", "avatars/1.png"))
	assert.True(t, fake.AssertCount("avatars", 2))
	assert.True(t, fake.AssertCount("", 0))
	assert.True(t, fake.AssertCount("", 2, true))
	assert.False(t, fake.AssertCount("missing", 0))
	assert.False(t, fake.AssertDirectoryEmpty(""))
	// The url of the disk is kept.
	assert.Equal(t, "https://example.com/storage/avatars/1.png", fake.Url("avatars/1.png"))

	assert.NoError(t, fake.MakeDirectory("empty"))
	assert.True(t, fake.AssertDirectoryEmpty("empty"))

	fakeRoot := fake.Path("")
	assert.NoError(t, fake.Reset())

	assert.NoDirExists(t, fakeRoot)
	assert.Same(t, original, storage.Driver)
	assert.Same(t, original, storage.Disk("local"))
}

func TestFake_DiskThatIsNotConfigured(t *testing.T) {
	storage, _ := newFakeTestStorage(t)

	fake := storage.Fake("s3")

	assert.Same(t, fake, storage.Disk("s3"))
	assert.NotSame(t, fake, storage.Driver)
	assert.NoError(t, storage.Disk("s3").Put("report.csv", "a,b"))
	assert.True(t, fake.AssertExists("report.csv"))

	assert.NoError(t, fake.Reset())

	// The disk was never loaded, so it is loaded again on its next use.
	storage.mu.RLock()
	_, loaded := storage.drivers["s3"]
	storage.mu.RUnlock()
	assert.False(t, loaded)
}

func TestFake_FakingAgainStartsEmpty(t *testing.T) {
	storage, _ := newFakeTestStorage(t)
	original := storage.Driver

	first := storage.Fake("local")
	assert.NoError(t, first.Put("a.txt", "a"))
	firstRoot := first.Path("")

	second := storage.Fake()

	assert.NoDirExists(t, firstRoot)
	assert.True(t, second.AssertMissing("a.txt"))
	assert.Same(t, second, storage.Driver)

	// Resetting the replaced fake changes nothing, the second one restores
	// the real disk.
	assert.NoError(t, first.Reset())
	assert.Same(t, second, storage.Driver)

	assert.NoError(t, second.Reset())
	assert.Same(t, original, storage.Driver)
}

func TestFake_ConcurrentDisk(t *testing.T) {
	storage, _ := newFakeTestStorage(t)
	fake := storage.Fake("s3")
	t.Cleanup(func() { _ = fake.Reset() })

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.Same(t, fake, storage.Disk("s3"))
		}()
	}
	wg.Wait()
}

func TestFake_RootIsATemporaryDirectory(t *testing.T) {
	storage, _ := newFakeTestStorage(t)
	fake := storage.Fake()
	t.Cleanup(func() { _ = fake.Reset() })

	info, err := os.Stat(fake.Path(""))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
