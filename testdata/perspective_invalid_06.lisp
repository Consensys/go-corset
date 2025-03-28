;;error:14:20-21:unknown symbol
;;error:15:16-17:unknown symbol
;;
;;
(defcolumns
    ;; Column (not in perspective)
    (A :i16)
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

(defperspective p1 P ((B :byte)))
(deflookup l1 (A) (B))
(deflookup l2 (B) (A))
