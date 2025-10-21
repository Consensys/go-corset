(defcolumns (X :i17) (P :binary) (Y :i16))
;; use of selector
(defclookup (test :unchecked) (Y) P (X))
