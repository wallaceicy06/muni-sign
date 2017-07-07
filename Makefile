SUBDIRS = admin driver nextbus

all: build
build: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir build; \
	done
clean: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir clean; \
	done
install: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir install; \
	done
test: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir test; \
	done
iref: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir iref; \
	done
fmt: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir test; \
	done

