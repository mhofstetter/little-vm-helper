package kernels

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	kernelImageRepository = "quay.io/lvh-images/kernel-images"

	kernelTagRegex = `^(.+)-([0-9]+\.[0-9]+|main)$`

	repoCommand = "repo"
	repoHelp    = "specify the OCI repository to pull the image from"
)

func catalogCommand() *cobra.Command {
	var (
		repo string
	)

	cmd := &cobra.Command{
		Use:   "catalog [version]",
		Short: "list available tags for kernel images from cilium/little-vm-helper-images",
		Long: `List the available tags for kernel images built from cilium/little-vm-helper-images

Examples:
  # List all available versions
  lvh kernels catalog

  # List the tags available for version 6.6
  lvh kernels catalog 6.6

  # Retrieve the latest tags available for version bpf-next
  lvh kernels catalog bpf-next | tail -n 2

  # Retrieve the latest CI-generated images for version bpf-next
  lvh kernels catalog bpf-next --repo quay.io/lvh-images/kernel-images-ci`,
		RunE: func(cmd *cobra.Command, args []string) error {
			re := regexp.MustCompile(kernelTagRegex)

			rawTagList, err := crane.ListTags(repo)
			if err != nil {
				return err
			}

			tags := map[string][]string{}
			for _, tag := range rawTagList {
				match := re.FindStringSubmatch(tag)
				if len(match) < 3 {
					// discard most of the tags that don't match the regex
					continue
				}

				if strings.Contains(match[1], "-latest") {
					// remove some tags with "-latest" that are obsolete
					continue
				}

				tags[match[1]] = append(tags[match[1]], match[0])
			}

			if len(args) == 0 {
				versions := []string{}
				for v := range tags {
					// semver package needs the v Prefix
					versions = append(versions, "v"+v)
				}
				semver.Sort(versions)
				for _, v := range versions {
					cmd.Println(strings.TrimPrefix(v, "v"))
				}
				return nil
			}

			if _, found := tags[args[0]]; !found {
				keys := []string{}
				for key := range tags {
					keys = append(keys, key)
				}
				return fmt.Errorf("kernel version not found, try: %s", keys)
			}

			for key := range tags {
				slices.Sort(tags[key])
			}
			for _, tag := range tags[args[0]] {
				cmd.Println(tag)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, repoCommand, kernelImageRepository, repoHelp)

	return cmd
}
