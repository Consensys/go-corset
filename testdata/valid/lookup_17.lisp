(defcolumns (X :i16) (Y :i16))
;; fragmented lookup with single fragment.
(defmlookup test ((Y)) ((X)))
