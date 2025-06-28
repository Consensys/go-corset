(defcolumns (P :binary) (X :i16) (Q :binary) (Y :i16))
;; use of selector
(defclookup test P (Y) Q (X))
