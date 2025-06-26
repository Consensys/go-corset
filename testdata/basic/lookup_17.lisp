(defcolumns (X :i16) (Y :i16))
;; fragmented lookup with single fragment.
(deflookup test ((Y)) (X))
