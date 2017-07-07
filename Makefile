SUBDIRS = proto admin driver nextbus

all: build
build: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir; \
	done
clean: $(SUBDIRS)
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir clean; \
	done
