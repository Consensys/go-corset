(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) ((shift X -1)))
