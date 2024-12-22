;;error:14:20-21:not visible here
;;error:15:16-17:not visible here
(defpurefun ((vanishes! :@loob :force) e0) e0)
;;
(defcolumns
    ;; Column (not in perspective)
    A
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

(defperspective p1 P ((B :byte)))
(deflookup l1 (A) (B))
(deflookup l2 (B) (A))
