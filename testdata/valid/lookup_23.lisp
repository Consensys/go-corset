(defcolumns (X :i16) (P :binary) (Y :i16))
;; use of selector
(defclookup test P (Y) 1 (X))
