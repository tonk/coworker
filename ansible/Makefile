COLLECTIONS_TMP := /tmp/warmdesk-molecule-collections

.PHONY: test lint setup clean-collections

# Create a symlink so Ansible can discover the collection under its namespace.
# ANSIBLE_COLLECTIONS_PATH is set to the temp dir; molecule.yml reads it.
setup:
	mkdir -p $(COLLECTIONS_TMP)/ansible_collections/ansilabnl
	ln -sfn $(shell pwd) $(COLLECTIONS_TMP)/ansible_collections/ansilabnl/warmdesk

# Run the full Molecule test suite.
test: setup
	molecule test -s default

# Run only converge (useful during development — leaves server running).
converge: setup
	molecule converge -s default

# Destroy test state (stop server, remove temp files).
destroy:
	molecule cleanup -s default

# Clean the collections temp dir.
clean-collections:
	rm -rf $(COLLECTIONS_TMP)

lint:
	ansible-lint plugins/modules/ plugins/lookup/ plugins/inventory/
