(defcolumns (X :i16) (P :binary) (Y :i16))
;; use of selector
(defclookup test (Y) P (X))
