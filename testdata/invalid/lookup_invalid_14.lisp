;;error:3:20-21:non-binary selector encountered
(defcolumns (P :u2) (X :u16) (Y :u16))
(defclookup l1 (X) P (Y))
