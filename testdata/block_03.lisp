(column X)
(column Y)
(column Z)
(vanish c1 (if X
               ;; if X==0 then Y == Z
               (begin Y Z)
               ;; else X == Y && (Y == 0 || Z == 0)
               (begin (- X Y) (* Y Z))))
;; Z is always 0!
(assert a1 Z)
